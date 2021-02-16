package handler

import (
	"context"

	common "github.com/baoshuai123/go-micro-common"
	"github.com/baoshuai123/go-micro-payment/domain/model"
	"github.com/baoshuai123/go-micro-payment/domain/service"
	payment "github.com/baoshuai123/go-micro-payment/proto/payment"
)

type Payment struct {
	PaymentDataService service.IPaymentDataService
}

//创建支付通道
func (p *Payment) AddPayment(ctx context.Context, request *payment.PaymentInfo, response *payment.PaymentID) error {
	paymentData := &model.Payment{}
	if err := common.SwapTo(request, paymentData); err != nil {
		common.Error(err)
	}
	paymentID, err := p.PaymentDataService.AddPayment(paymentData)
	if err != nil {
		common.Error(err)
	}
	response.PaymentId = paymentID
	return nil
}

// 更新支付通道
func (p *Payment) UpdatePayment(ctx context.Context, request *payment.PaymentInfo, response *payment.Response) error {
	paymentData := &model.Payment{}
	if err := common.SwapTo(request, paymentData); err != nil {
		common.Error(err)
	}
	err := p.PaymentDataService.UpdatePayment(paymentData)
	if err != nil {
		common.Error(err)
	}
	response.Msg = "更新成功"
	return nil
}

// 根据id删除支付通道
func (p *Payment) DeletePaymentByID(ctx context.Context, request *payment.PaymentID, response *payment.Response) error {
	return p.PaymentDataService.DeletePayment(request.PaymentId)
}

// 根据id查找支付通道
func (p *Payment) FindPaymentByID(ctx context.Context, request *payment.PaymentID, response *payment.PaymentInfo) error {
	paymentData, err := p.PaymentDataService.FindPaymentByID(request.PaymentId)
	if err != nil {
		common.Error(err)
	}
	return common.SwapTo(paymentData, response)
}

//查找所有支付通道
func (p *Payment) FindAllPayment(ctx context.Context, request *payment.All, response *payment.PaymentAll) error {
	allPayment, err := p.PaymentDataService.FindAllPayment()
	if err != nil {
		common.Error(err)
	}
	for _, v := range allPayment {
		paymentInfo := &payment.PaymentInfo{}
		if err := common.SwapTo(v, paymentInfo); err != nil {
			common.Error(err)
		}
		response.PaymentInfo = append(response.PaymentInfo, paymentInfo)
	}
	return nil
}
