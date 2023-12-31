// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/adrianbrad/privatebtc"
	"sync"
)

// Ensure, that NodeService does implement privatebtc.NodeService.
// If this is not the case, regenerate this file with moq.
var _ privatebtc.NodeService = &NodeService{}

// NodeService is a mock implementation of privatebtc.NodeService.
//
//	func TestSomethingThatUsesNodeService(t *testing.T) {
//
//		// make and configure a mocked privatebtc.NodeService
//		mockedNodeService := &NodeService{
//			CreateNodesFunc: func(ctx context.Context, nodeRequests []privatebtc.CreateNodeRequest) ([]privatebtc.NodeHandler, error) {
//				panic("mock out the CreateNodes method")
//			},
//		}
//
//		// use mockedNodeService in code that requires privatebtc.NodeService
//		// and then make assertions.
//
//	}
type NodeService struct {
	// CreateNodesFunc mocks the CreateNodes method.
	CreateNodesFunc func(ctx context.Context, nodeRequests []privatebtc.CreateNodeRequest) ([]privatebtc.NodeHandler, error)

	// calls tracks calls to the methods.
	calls struct {
		// CreateNodes holds details about calls to the CreateNodes method.
		CreateNodes []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// NodeRequests is the nodeRequests argument value.
			NodeRequests []privatebtc.CreateNodeRequest
		}
	}
	lockCreateNodes sync.RWMutex
}

// CreateNodes calls CreateNodesFunc.
func (mock *NodeService) CreateNodes(ctx context.Context, nodeRequests []privatebtc.CreateNodeRequest) ([]privatebtc.NodeHandler, error) {
	if mock.CreateNodesFunc == nil {
		panic("NodeService.CreateNodesFunc: method is nil but NodeService.CreateNodes was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		NodeRequests []privatebtc.CreateNodeRequest
	}{
		Ctx:          ctx,
		NodeRequests: nodeRequests,
	}
	mock.lockCreateNodes.Lock()
	mock.calls.CreateNodes = append(mock.calls.CreateNodes, callInfo)
	mock.lockCreateNodes.Unlock()
	return mock.CreateNodesFunc(ctx, nodeRequests)
}

// CreateNodesCalls gets all the calls that were made to CreateNodes.
// Check the length with:
//
//	len(mockedNodeService.CreateNodesCalls())
func (mock *NodeService) CreateNodesCalls() []struct {
	Ctx          context.Context
	NodeRequests []privatebtc.CreateNodeRequest
} {
	var calls []struct {
		Ctx          context.Context
		NodeRequests []privatebtc.CreateNodeRequest
	}
	mock.lockCreateNodes.RLock()
	calls = mock.calls.CreateNodes
	mock.lockCreateNodes.RUnlock()
	return calls
}
