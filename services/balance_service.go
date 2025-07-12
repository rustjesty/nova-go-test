package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"solana-balance-api/cache"
	"solana-balance-api/models"
)

type BalanceService struct {
	Client      *rpc.Client
	Cache       *cache.Cache
	Mutexes     map[string]*sync.Mutex
	MutexesLock sync.RWMutex
}

func NewBalanceService(client *rpc.Client, cache *cache.Cache) *BalanceService {
	return &BalanceService{
		Client:  client,
		Cache:   cache,
		Mutexes: make(map[string]*sync.Mutex),
	}
}

func (s *BalanceService) getAddressMutex(address string) *sync.Mutex {
	s.MutexesLock.Lock()
	defer s.MutexesLock.Unlock()
	if m, exists := s.Mutexes[address]; exists {
		return m
	}
	m := &sync.Mutex{}
	s.Mutexes[address] = m
	return m
}

func (s *BalanceService) FetchBalanceFromSolana(address string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, fmt.Errorf("invalid address: %v", err)
	}
	balance, err := s.Client.GetBalance(ctx, pubKey, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %v", err)
	}
	return float64(balance.Value) / 1_000_000_000, nil
}

func (s *BalanceService) ProcessAddress(address string) models.BalanceItem {
	mutex := s.getAddressMutex(address)
	mutex.Lock()
	defer mutex.Unlock()
	if balance, cached := s.Cache.Get(address); cached {
		return models.BalanceItem{
			Address: address,
			Balance: balance,
		}
	}
	balance, err := s.FetchBalanceFromSolana(address)
	if err != nil {
		return models.BalanceItem{
			Address: address,
			Error:   err.Error(),
		}
	}
	s.Cache.Set(address, balance)
	return models.BalanceItem{
		Address: address,
		Balance: balance,
	}
} 