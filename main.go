package main

import (
	"fmt"
	"log"
	"occlum-securemem/securemem"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	fmt.Println("Secure Memory Vault Demo (SGX + Occlum + Go)")

	vault, err := securemem.NewMemoryVault()
	if err != nil {
		log.Fatalf("Failed to create vault: %v", err)
	}

	// 模拟加密保存结构体
	user := User{ID: 42, Name: "Alice", Email: "alice@example.com"}
	if err := vault.Put("user1", user); err != nil {
		log.Fatalf("Put failed: %v", err)
	}
	fmt.Println("User encrypted and stored in memory")

	var recovered User
	if err := vault.Get("user1", &recovered); err != nil {
		log.Fatalf("Get failed: %v", err)
	}
	fmt.Printf("Decrypted User: %+v\n", recovered)
}
