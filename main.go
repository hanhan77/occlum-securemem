package main

import (
	"fmt"
	"occlum-securemem/securemem"
	"os"
)

type User struct {
	ID    int
	Name  string
	Email string
}

func main() {
	fmt.Println("Secure Memory Vault Persistent Demo (SGX + Occlum + Go)")

	const persistPath = "/data-secure/vault.bin"

	// 创建或加载 vault
	var vault *securemem.MemoryVault
	if _, err := os.Stat(persistPath); err == nil {
		// 落盘已存在，加载
		vault, _ = securemem.NewMemoryVault()
		if err := vault.LoadFromFile(persistPath); err != nil {
			fmt.Println("Load failed:", err)
			return
		}
		fmt.Println("Vault restored from disk.")
	} else {
		// 初次运行，创建并添加数据
		vault, _ = securemem.NewMemoryVault()
		vault.Put("user_42", &User{ID: 42, Name: "Alice", Email: "alice@example.com"})
		if err := vault.PersistToFile(persistPath); err != nil {
			fmt.Println("Persist failed:", err)
			return
		}
		fmt.Println("Vault created and persisted.")
	}

	// 读取数据
	var u User
	if err := vault.Get("user_42", &u); err != nil {
		fmt.Println("Get failed:", err)
		return
	}
	fmt.Printf("Decrypted User: %+v\n", u)
}
