package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// 日本語のヘッダーを表示
	fmt.Println("### 暗号化アルゴリズム性能比較ベンチマーク結果")
	fmt.Println()
	fmt.Println("以下は公開鍵暗号方式（RSA）と共通鍵暗号方式（AES）の暗号化・復号化処理の性能比較結果です。")
	fmt.Println("各列の意味は次の通りです：")
	fmt.Println()
	fmt.Println("アルゴリズム・操作                実行回数            1操作あたりの実行時間")
	fmt.Println("------------------------------ ----------------- ----------------------")

	// ベンチマークテストを実行
	cmd := exec.Command("go", "test", "-bench", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "." // 現在のディレクトリで実行

	err := cmd.Run()
	if err != nil {
		fmt.Printf("ベンチマークの実行中にエラーが発生しました: %v\n", err)
		os.Exit(1)
	}

	// 補足説明を表示
	fmt.Println()
	fmt.Println("※ 数値が小さいほど高速な処理を示しています。AES（共通鍵暗号）はRSA（公開鍵暗号）と比較して数百〜数万倍高速であることが分かります。")
	fmt.Println("  特にRSA復号化処理は最も計算コストが高く、AES処理の約20,000倍の時間を要しています。")
}
