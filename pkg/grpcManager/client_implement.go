// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

package grpcManager

import (
	"context"
	"fmt"
	"github.com/elf-io/elf/api/v1/grpcService"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *grpcClientManager) SendRequestForExecRequest(ctx context.Context, serverAddress []string, requestList []*grpcService.ExecRequestMsg) ([]*grpcService.ExecResponseMsg, error) {

	logger := s.logger.With(
		zap.String("server", fmt.Sprintf("%v", serverAddress)),
	)

	if e := s.clientDial(ctx, serverAddress); e != nil {
		return nil, errors.Errorf("failed to dial, error=%v", e)
	}
	defer s.client.Close()

	response := []*grpcService.ExecResponseMsg{}

	c := grpcService.NewCmdServiceClient(s.client)
	stream, err := c.ExecRemoteCmd(ctx)
	if err != nil {
		return nil, err
	}

	for n, request := range requestList {
		logger.Sugar().Debugf("send %v request ", n)
		if err := stream.Send(request); err != nil {
			return nil, err
		}

		if r, err := stream.Recv(); err != nil {
			return nil, err
		} else {
			logger.Sugar().Debugf("recv %v response ", n)
			response = append(response, r)
		}
	}

	logger.Debug("finish")
	if e := stream.CloseSend(); e != nil {
		logger.Sugar().Errorf("grpc failed to CloseSend error=%v ", e)
	}
	return response, nil

}
