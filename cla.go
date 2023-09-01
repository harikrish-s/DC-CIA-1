package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Process struct {
	ID int
	Accounts [3]int
	Channel chan Transaction
	LocalSnapshot []int
	State []int
}

type Transaction struct {
	From int
	To int
	Amount int
}

var (
	no_processes int
	processes []*Process
	wg sync.WaitGroup
)

func main() {
	fmt.Print("Enter the number of processes: ")
	fmt.Scan(&no_processes)
	rand.Seed(time.Now().UnixNano())
	initializeProcesses()
	fmt.Println("Initial total amount:", get_tot_amt())

	for _, p := range processes {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			for i := 0; i < no_processes*15; i++ {
				transaction := generateTransaction(proc.ID)
				proc.transaction(transaction)
				if i > 0 && i%no_processes == 0 {
					proc.initiateSnapshot()
				}
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			}
		}(p)
	}

	wg.Wait()
	fmt.Println("Final amount:", get_tot_amt())
}

func initializeProcesses() {
	processes = make([]*Process, no_processes)
	for i := 0; i < no_processes; i++ {
		processes[i] = &Process{
			ID: i,
			Channel: make(chan Transaction),
			Accounts: [3]int{
				rand.Intn(10000),
				rand.Intn(10000),
				rand.Intn(10000),
			},
			LocalSnapshot: make([]int, no_processes),
			State: make([]int, no_processes),
		}
	}
}

func get_tot_amt() int {
	total := 0
	for _, p := range processes {
		for _, account := range p.Accounts {
			total += account
		}
	}
	return total
}

func generateTransaction(senderID int) Transaction {
	from := senderID
	to := rand.Intn(no_processes)
	amount := rand.Intn(100)
	return Transaction{From: from, To: to, Amount: amount}
}

func (p *Process) transaction(t Transaction) {
	p.Accounts[t.From] -= t.Amount
	p.Accounts[t.To] += t.Amount
	fmt.Printf("Process %d: Sent %d from account %d to account %d\n", p.ID, t.Amount, t.From, t.To)
}

func (p *Process) initiateSnapshot() {
	fmt.Printf("Process %d initiating snapshot\n", p.ID)
	for i := 0; i < no_processes; i++ {
		p.LocalSnapshot[i] = p.Accounts[i]
		p.State[i] = p.Accounts[i]
	}

	for _, proc := range processes {
		if proc.ID != p.ID {
			go func(receiver *Process) {
				fmt.Printf("Process %d: Sending marker to process %d\n", p.ID, receiver.ID)
				receiver.Channel <- Transaction{
					From:   p.ID,
					To:     receiver.ID,
					Amount: 0,
				}
			}(proc)
		}
	}

	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Snapshot taken by process %d: %v\n", p.ID, p.LocalSnapshot)
}