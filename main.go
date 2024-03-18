package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/go-pdf/fpdf"
)

type Buku struct {
	KodeBuku      string
	JudulBuku     string
	PengarangBuku string
	PenerbitBuku  string
	JumlahHalaman int
	TahunTerbit   int
	Waktu         time.Time
}

var ListBuku []Buku

func TambahBuku() {
	BookCode := ""
	BookTitle := ""
	BookAuthor := ""
	BookPublisher := ""
	var PageTotal int
	var PublishedYear int

	// fmt.Print("\n")
	fmt.Println("\n===== Tambah Buku =====")

	draftBuku := []Buku{}

	for {
		fmt.Print("Masukkan Kode Buku : ")
		_, err := fmt.Scanln(&BookCode)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		for _, book := range draftBuku {
			if book.KodeBuku == "book-"+BookCode {
				fmt.Println("Kode Buku Sudah Dipakai. Masukkan Kode Buku Lain")
				return
			}
		}

		listJsonBuku, err := os.ReadDir("books")
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}
		for _, bookJson := range listJsonBuku {
			if bookJson.Name() == "book-"+BookCode+".json" {
				fmt.Println("Kode Buku Sudah Dipakai. Masukkan Kode Buku Lain.")
				return
			}
		}

		fmt.Print("Masukkan Judul Buku : ")
		_, err = fmt.Scanln(&BookTitle)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		fmt.Print("Masukkan Pengarang Buku : ")
		_, err = fmt.Scanln(&BookAuthor)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		fmt.Print("Masukkan Penerbit Buku : ")
		_, err = fmt.Scanln(&BookPublisher)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		fmt.Print("Masukkan Jumlah Halaman Buku : ")
		_, err = fmt.Scanln(&PageTotal)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		fmt.Print("Masukkan Tahun Terbit Buku : ")
		_, err = fmt.Scanln(&PublishedYear)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		draftBuku = append(draftBuku, Buku{
			KodeBuku:      fmt.Sprintf("book-%s", BookCode),
			JudulBuku:     BookTitle,
			PengarangBuku: BookAuthor,
			PenerbitBuku:  BookPublisher,
			JumlahHalaman: PageTotal,
			TahunTerbit:   PublishedYear,
			Waktu:         time.Now(),
		})

		pilihanTambahBuku := 0
		fmt.Println("Ketik 1 untuk tambah buku lain, ketik 0 untuk selesai")
		_, err = fmt.Scanln(&pilihanTambahBuku)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		if pilihanTambahBuku == 0 {
			break
		}
	}

	fmt.Println("Menambah Buku...")

	_ = os.Mkdir("books", 0777)

	ch := make(chan Buku)

	wg := sync.WaitGroup{}

	jumlahPekerja := 5

	// Mendaftarkan receiver/pemroses data
	for i := 0; i < jumlahPekerja; i++ {
		wg.Add(1)
		go SimpanBuku(ch, &wg, i)
	}

	for _, bukuTersimpan := range draftBuku {
		ch <- bukuTersimpan
	}

	close(ch)

	wg.Wait()

	fmt.Println("Berhasil Menambah Buku")

	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func SimpanBuku(ch <-chan Buku, wg *sync.WaitGroup, noPekerja int) {

	for bukuTersimpan := range ch {
		dataJson, err := json.Marshal(bukuTersimpan)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		err = os.WriteFile(fmt.Sprintf("books/%s.json", bukuTersimpan.KodeBuku), dataJson, 0644)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}

		fmt.Printf("Pekerja No %d Memproses Kode Buku : %s!\n", noPekerja, bukuTersimpan.KodeBuku)
	}
	wg.Done()

	// bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func LihatDaftarBuku(ch <-chan string, chBuku chan Buku, wg *sync.WaitGroup) {
	var buku Buku
	for KodeBuku := range ch {
		dataJson, err := os.ReadFile(fmt.Sprintf("books/%s", KodeBuku))
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
		}

		err = json.Unmarshal(dataJson, &buku)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
		}

		chBuku <- buku
	}
	wg.Done()

	// bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func DaftarBuku() {
	// fmt.Print("\n")

	fmt.Println("===== Daftar Buku =====")
	fmt.Println("Memuat Data...")
	ListBuku = []Buku{}

	listJsonBuku, err := os.ReadDir("books")
	if err != nil {
		fmt.Println("Terjadi Error: ", err)
	}

	wg := sync.WaitGroup{}

	ch := make(chan string)
	chBuku := make(chan Buku, len(listJsonBuku))

	jumlahPekerja := 5

	for i := 0; i < jumlahPekerja; i++ {
		wg.Add(1)
		go LihatDaftarBuku(ch, chBuku, &wg)
	}

	for _, fileBuku := range listJsonBuku {
		ch <- fileBuku.Name()
	}

	close(ch)

	wg.Wait()

	close(chBuku)

	for dataBuku := range chBuku {
		ListBuku = append(ListBuku, dataBuku)
	}

	sort.Slice(ListBuku, func(i, j int) bool {
		return ListBuku[i].Waktu.Before(ListBuku[j].Waktu)
	})

	if len(ListBuku) < 1 {
		fmt.Println("----- Tidak Ada Buku -----")
	}

	for i, v := range ListBuku {
		i++
		fmt.Printf("%d. Kode Buku : %s, Judul Buku : %s, Pengarang : %s, Penerbit : %s, Jumhlah Halaman : %d, Tahun Terbit : %d\n",
			i, v.KodeBuku, v.JudulBuku, v.PengarangBuku, v.PenerbitBuku, v.JumlahHalaman, v.TahunTerbit)
	}

	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func DetailBuku(kode string) {
	// fmt.Print("\n")
	fmt.Println("===== Detail Buku =====")

	var isiBuku bool

	for _, buku := range ListBuku {
		if buku.KodeBuku == kode {
			isiBuku = true
			fmt.Printf("Kode Buku : %s\n", buku.KodeBuku)
			fmt.Printf("Judul Buku : %s\n", buku.JudulBuku)
			fmt.Printf("Pengarang Buku : %s\n", buku.PengarangBuku)
			fmt.Printf("Penerbit Buku : %s\n", buku.PenerbitBuku)
			fmt.Printf("Jumlah Halaman Buku : %d\n", buku.JumlahHalaman)
			fmt.Printf("Tahun Terbit Buku : %d\n", buku.TahunTerbit)
			break
		}
	}

	if !isiBuku {
		fmt.Println("Kode Buku Salah Atau Tidak Ada")
	}

	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func EditBuku(kode string) {

	DetailBuku(kode)

	// fmt.Print("\n")
	fmt.Println("===== Edit Buku =====")

	var buku Buku

	fmt.Print("Masukkan Kode Buku : ")
	_, err := fmt.Scanln(&buku.KodeBuku)
	if err != nil {
		fmt.Println("Terjadi Error: ", err)
	}

	listJsonBuku, err := os.ReadDir("books")
	if err != nil {
		fmt.Println("Terjadi Error: ", err)
	}

	for _, bukuJson := range listJsonBuku {
		if bukuJson.Name() == "book-"+buku.KodeBuku+".json" {
			fmt.Println("Kode Buku Sudah Ada. Masukkan Kode Buku Lain")
			return
		}
	}

	fmt.Print("Masukkan Judul Buku : ")
	_, err = fmt.Scanln(&buku.JudulBuku)
	if err != nil {
		fmt.Scanln("Terjadi Error : ", err)
		return
	}

	fmt.Print("Masukkan Pengarang Buku : ")
	_, err = fmt.Scanln(&buku.PengarangBuku)
	if err != nil {
		fmt.Scanln("Terjadi Error : ", err)
		return
	}

	fmt.Print("Masukkan Penerbit Buku : ")
	_, err = fmt.Scanln(&buku.PenerbitBuku)
	if err != nil {
		fmt.Scanln("Terjadi Error : ", err)
		return
	}

	fmt.Print("Masukkan Jumlah Halaman Buku : ")
	_, err = fmt.Scanln(&buku.JumlahHalaman)
	if err != nil {
		fmt.Scanln("Terjadi Error : ", err)
		return
	}

	fmt.Print("Masukkan Tahun Terbit Buku : ")
	_, err = fmt.Scanln(&buku.TahunTerbit)
	if err != nil {
		fmt.Scanln("Terjadi Error : ", err)
		return
	}
	fmt.Println("\nBuku Berhasil Diedit!")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	buku.KodeBuku = "book-" + buku.KodeBuku
	fmt.Println(buku)

	for i, b := range ListBuku {
		if b.KodeBuku == kode {
			ListBuku[i] = buku
			dataJson, err := json.Marshal(ListBuku[i])
			if err != nil {
				fmt.Println("Terjadi Error: ", err)
			}

			err = os.WriteFile(fmt.Sprintf("books/%s.json", ListBuku[i].KodeBuku), dataJson, 0644)
			if err != nil {
				fmt.Println("Terjadi Error: ", err)
			}

			err = os.Remove(fmt.Sprintf("books/%s.json", kode))
			if err != nil {
				fmt.Println("Terjadi Error: ", err)
			}

			break
		}
	}

	// bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func HapusBuku(kode string) {

	var isiBuku bool
	for i, buku := range ListBuku {
		if buku.KodeBuku == kode {
			isiBuku = true
			err := os.Remove(fmt.Sprintf("books/%s.json", ListBuku[i].KodeBuku))
			if err != nil {
				fmt.Println("Terjadi Error: ", err)
			}

			fmt.Print("\n")
			fmt.Println("Buku Berhasil Dihapus")
			break
		}
	}

	if !isiBuku {

		fmt.Print("\n")
		fmt.Println("Kode Buku Salah Atau Tidak Ada")
	}

	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func GeneratePdfBuku() {

	// fmt.Print("\n")
	DaftarBuku()
	fmt.Println("===== Membuat Daftar Buku... =====")
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "", 12)
	pdf.SetLeftMargin(10)
	pdf.SetRightMargin(10)

	for i, buku := range ListBuku {
		bukuText := fmt.Sprintf(
			"Buku #%d:\nKode Buku : %s\nJudul : %s\nPengarang : %s\nPenerbit : %s\nJumlah Halaman : %d\nTahun Terbit : %d\nWaktu : %s\n",
			i+1, buku.KodeBuku, buku.JudulBuku, buku.PengarangBuku, buku.PenerbitBuku, buku.JumlahHalaman, buku.TahunTerbit,
			buku.Waktu.Format("2006-01-02 15:04:05"))

		pdf.MultiCell(0, 10, bukuText, "0", "L", false)
		pdf.Ln(5)
	}

	err := pdf.OutputFileAndClose(
		fmt.Sprintf("daftar_buku_%s.pdf",
			time.Now().Format("2006-01-02 15-04-05")))

	if err != nil {
		fmt.Println("Terjadi Error : ", err)
	}

	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func PrintSelectedBook() {
	DaftarBuku()
	fmt.Print("Masukkan nomor urut buku yang ingin dicetak: ")
	var selectedNumber int
	_, err := fmt.Scanln(&selectedNumber)
	if err != nil {
		fmt.Println("Terjadi error:", err)
		return
	}
	if selectedNumber < 1 || selectedNumber > len(ListBuku) {
		fmt.Println("Nomor urut buku tidak valid.")
		return
	}
	selectedBook := ListBuku[selectedNumber-1]
	Selected(selectedBook)
}

func Selected(selectedBook Buku) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.SetLeftMargin(10)
	pdf.SetRightMargin(10)

	bukuText := fmt.Sprintf(
		"====================================\nKodeBuku : %s\nJudulBuku : %s\nPengarang : %s\nPenerbit : %s\nJumlahHalaman : %d\nTahunTerbit :  %d\nTanggal : %s\n====================================\n",
		selectedBook.KodeBuku, selectedBook.JudulBuku, selectedBook.PengarangBuku, selectedBook.PenerbitBuku, selectedBook.JumlahHalaman, selectedBook.TahunTerbit,
		selectedBook.Waktu.Format("2006-01-02 15:04:05"))

	pdf.MultiCell(0, 10, bukuText, "0", "L", false)

	err := pdf.OutputFileAndClose(
		fmt.Sprintf("data_buku_%s.pdf",
			time.Now().Format("2006-01-02-15-04-05")))

	if err != nil {
		fmt.Println("Terjadi error:", err)
	}
	fmt.Println("Buku berhasil dicetak dalam file PDF.")
}

func main() {

	var pilihan int

	fmt.Println("\n---==== Perpustakaan Desa Sukatidur ====---")
	fmt.Println("---==== Manajemen Buku Perpustakaan ====---")

	fmt.Println("Pilih Opsi")
	fmt.Println("1. Tambah Buku")
	fmt.Println("2. Daftar Buku")
	fmt.Println("3. Detail Buku")
	fmt.Println("4. Edit Buku")
	fmt.Println("5. Hapus Buku")
	fmt.Println("6. Generate PDF")
	fmt.Println("7. Keluar")

	fmt.Print("Masukkan Opsi : ")
	_, err := fmt.Scanln(&pilihan)
	if err != nil {
		fmt.Println("Terjadi Error: ", err)
		return
	}

	switch pilihan {
	case 1:
		TambahBuku()
	case 2:
		DaftarBuku()
	case 3:
		var pilihanDetail string
		DaftarBuku()
		fmt.Print("Masukkan Kode Buku : ")
		_, err := fmt.Scanln(&pilihanDetail)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}
		DetailBuku(pilihanDetail)
	case 4:
		var pilihanEdit string
		DaftarBuku()
		fmt.Print("Masukkan Kode Buku Yang Akan Diedit : ")
		_, err := fmt.Scanln(&pilihanEdit)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}
		EditBuku(pilihanEdit)
	case 5:
		var pilihanHapus string
		DaftarBuku()
		fmt.Print("Masukkan Kode Buku Yang Akan Dihapus : ")
		_, err := fmt.Scanln(&pilihanHapus)
		if err != nil {
			fmt.Println("Terjadi Error: ", err)
			return
		}
		HapusBuku(pilihanHapus)
	case 6:
		GeneratePdfBuku()
	case 7:
		os.Exit(0)
	default:
		fmt.Println("Tidak Ada Opsi")
	}

	main()
}
