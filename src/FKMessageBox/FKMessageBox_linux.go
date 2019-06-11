package FKMessageBox

import "fmt"

func MessageBox_Notice(title, caption string) {
	fmt.Println("%s : %s", title, caption)
}

func MessageBox_OkCancel(title, caption string) int {
	fmt.Println("%s : %s", title, caption)
	return 0
}
