package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	v1 "snowflake/api/snowflake/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/glog"
)

const (
	epoch             = int64(1577808000000)                           // 设置起始时间(时间戳/毫秒)：2020-01-01 00:00:00，有效期69年
	timestampBits     = uint(41)                                       // 时间戳占用位数
	datacenteridBits  = uint(5)                                        // 数据中心id所占位数
	workeridBits      = uint(5)                                        // 机器id所占位数
	sequenceBits      = uint(12)                                       // 序列所占的位数
	timestampMax      = int64(-1 ^ (-1 << timestampBits))              // 时间戳最大值
	datacenteridMax   = int64(-1 ^ (-1 << datacenteridBits))           // 支持的最大数据中心id数量
	workeridMax       = int64(-1 ^ (-1 << workeridBits))               // 支持的最大机器id数量
	sequenceMask      = int64(-1 ^ (-1 << sequenceBits))               // 支持的最大序列id数量
	workeridShift     = sequenceBits                                   // 机器id左移位数
	datacenteridShift = sequenceBits + workeridBits                    // 数据中心id左移位数
	timestampShift    = sequenceBits + workeridBits + datacenteridBits // 时间戳左移位数
)

type Snowflake struct {
	sync.Mutex
	timestamp    int64
	workerId     int64
	dataCenterId int64
	sequence     int64
}

type SnowflakeService struct {
	v1.UnimplementedSnowflakeServer

	log *log.Helper
}

func NewSnowflakeService(logger log.Logger) *SnowflakeService {
	return &SnowflakeService{log: log.NewHelper(logger)}
}

func (s *SnowflakeService) CreateSnowflake(ctx context.Context, req *v1.CreateSnowflakeRequest) (*v1.CreateSnowflakeReply, error) {
	s.log.WithContext(ctx).Infof("dataCenterId Received: %d workerId Received: %d", req.GetDataCenterId(), req.GetWorkerId())

	sn, err := NewSnowflake(req.GetDataCenterId(), req.GetWorkerId())
	if err != nil {
		return nil, err
	}

	return &v1.CreateSnowflakeReply{
		SnowFlakeId: sn.NextVal(),
	}, nil
}

func NewSnowflake(dataCenterId, workerId int64) (*Snowflake, error) {
	if dataCenterId < 0 || dataCenterId > datacenteridMax {
		return nil, fmt.Errorf("dataCenterId must be between 0 and %d", datacenteridMax-1)
	}
	if workerId < 0 || workerId > workeridMax {
		return nil, fmt.Errorf("workerId must be between 0 and %d", workeridMax-1)
	}
	return &Snowflake{
		timestamp:    0,
		dataCenterId: dataCenterId,
		workerId:     workerId,
		sequence:     0,
	}, nil
}

func (s *Snowflake) NextVal() int64 {
	s.Lock()
	now := time.Now().UnixNano() / 1000000 // 转毫秒
	if s.timestamp == now {
		// 当同一时间戳（精度：毫秒）下多次生成id会增加序列号
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 如果当前序列超出12bit长度，则需要等待下一毫秒
			// 下一毫秒将使用sequence:0
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		// 不同时间戳（精度：毫秒）下直接使用序列号：0
		s.sequence = 0
	}
	t := now - epoch
	if t > timestampMax {
		s.Unlock()
		glog.Errorf("epoch must be between 0 and %d", timestampMax-1)
		return 0
	}
	s.timestamp = now
	r := int64((t)<<timestampShift | (s.dataCenterId << datacenteridShift) | (s.workerId << workeridShift) | (s.sequence))
	s.Unlock()
	return r
}

// GetDeviceID 获取数据中心ID和机器ID
func GetDeviceID(sid int64) (dataCenterId, workerId int64) {
	dataCenterId = (sid >> datacenteridShift) & datacenteridMax
	workerId = (sid >> workeridShift) & workeridMax
	return
}

// GetTimestamp 获取时间戳
func GetTimestamp(sid int64) (timestamp int64) {
	timestamp = (sid >> timestampShift) & timestampMax
	return
}

// GetGenTimestamp 获取创建ID时的时间戳
func GetGenTimestamp(sid int64) (timestamp int64) {
	timestamp = GetTimestamp(sid) + epoch
	return
}

// GetGenTime 获取创建ID时的时间字符串(精度：秒)
func GetGenTime(sid int64) (t string) {
	// 需将GetGenTimestamp获取的时间戳/1000转换成秒
	t = time.Unix(GetGenTimestamp(sid)/1000, 0).Format("2006-01-02 15:04:05")
	return
}

// GetTimestampStatus 获取时间戳已使用的占比：范围（0.0 - 1.0）
func GetTimestampStatus() (state float64) {
	state = float64(time.Now().UnixNano()/1000000-epoch) / float64(timestampMax)
	return
}
