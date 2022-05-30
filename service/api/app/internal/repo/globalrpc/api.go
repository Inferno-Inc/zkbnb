package globalrpc

import (
	"context"
	"sync"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalrpc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GlobalRPC interface {
	GetLatestAccountInfo(accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
	GetLatestL1Amount(assetId uint16) (totalAmount int64, err error)
	GetLatestL1AmountList() (amounts []*types.AmountInfo, err error)
	GetSwapAmount(pairIndex, assetId uint16, assetAmount uint64, isFrom bool) (assetAmount uint64, pairIndex, assetId uint16)
	GetLatestTxsListByAccountIndexAndTxType(accountIndex uint64, txType uint64, limit uint64, offset uint64) ([]*mempool.MempoolTx, error)
}

var singletonValue *globalRPC
var once sync.Once
var c config.Config

func New(c config.Config, ctx context.Context) GlobalRPC {
	once.Do(func() {
		conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
		gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
		if err != nil {
			logx.Errorf("gorm connect db error, err = %s", err.Error())
		}
		redisConn := redis.New(c.CacheRedis[0].Host, func(p *redis.Redis) {
			p.Type = c.CacheRedis[0].Type
			p.Pass = c.CacheRedis[0].Pass
		})
		singletonValue = &globalRPC{
			AccountModel:        account.NewAccountModel(conn, c.CacheRedis, gormPointer),
			AccountHistoryModel: account.NewAccountHistoryModel(conn, c.CacheRedis, gormPointer),
			MempoolModel:        mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
			MempoolDetailModel:  mempool.NewMempoolDetailModel(conn, c.CacheRedis, gormPointer),
			RedisConnection:     redisConn,
			globalRPC:           globalrpc.NewGlobalRPC(zrpc.MustNewClient(c.GlobalRpc)),
			ctx:                 ctx,
		}
	})
	return singletonValue
}
