package main

import (
	"fmt"

	"github.com/baoshuai123/go-micro-payment/domain/repository"
	service2 "github.com/baoshuai123/go-micro-payment/domain/service"
	"github.com/baoshuai123/go-micro-payment/handler"

	"github.com/micro/go-plugins/wrapper/monitoring/prometheus/v2"

	ratelimit "github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2"

	opentracing2 "github.com/micro/go-plugins/wrapper/trace/opentracing/v2"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/opentracing/opentracing-go"

	"github.com/micro/go-micro/v2/registry"
	consul2 "github.com/micro/go-plugins/registry/consul/v2"

	payment "github.com/baoshuai123/go-micro-payment/proto/payment"

	common "github.com/baoshuai123/go-micro-common"
	"github.com/micro/go-micro/v2"
	log "github.com/micro/go-micro/v2/logger"
	"gorm.io/driver/mysql"
)

func main() {
	//配置中心
	consulConfig, err := common.GetConsulConfig("localhost", 8500, "/micro/config")
	if err != nil {
		common.Error(err)
	}
	//注册中心
	consul := consul2.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{
			"localhost:8500",
		}
	})
	//链路追踪
	tracer, io, err := common.NewTracer("go.micro.service.payment", "localhost:6831")
	if err != nil {
		common.Error(err)
	}
	defer io.Close()
	opentracing.SetGlobalTracer(tracer)
	//mysql 设置
	mysqlInfo := common.GetMysqlFromConsul(consulConfig, "mysql")
	//初始化数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlInfo.User,
		mysqlInfo.Pwd,
		mysqlInfo.Host,
		mysqlInfo.Port,
		mysqlInfo.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		common.Error(err)
	}
	//初始化表
	tableInit := repository.NewPaymentRepository(db)
	if err := tableInit.InitTable(); err != nil {
		common.Error(err)
	}
	//监控
	common.PrometheusBoot(9089)

	// New Service
	service := micro.NewService(
		micro.Name("go.micro.service.payment"),
		micro.Version("latest"),
		micro.Address("0.0.0.0:8089"),
		//添加注册中心
		micro.Registry(consul),
		//添加链路追踪
		micro.WrapHandler(opentracing2.NewHandlerWrapper(opentracing.GlobalTracer())),
		//加载限流
		micro.WrapHandler(ratelimit.NewHandlerWrapper(1000)),
		//加载监控
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
	)

	// Initialise service
	service.Init()

	paymentDataService := service2.NewPaymentDataService(tableInit)

	// Register Handler
	payment.RegisterPaymentHandler(service.Server(), &handler.Payment{PaymentDataService: paymentDataService})

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
