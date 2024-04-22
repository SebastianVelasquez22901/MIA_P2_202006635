package Comandos

import (
	"Proyecto/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/fatih/color"
)

func Comparar(a string, b string) bool {
	return strings.EqualFold(a, b)
}
func MayorQue(a int, b int) bool {
	return a > b
}
func Suma(a int, b int) int {
	return a + b
}

func Resta(a int, b int) int {
	return a - b
}

func MenorQue(a int, b int) bool {
	return a < b
}

func Error(op string, mensaje string) {
	color.Red("\tERROR: " + op + "\n\tTIPO: " + mensaje)
}

func Mensaje(op string, mensaje string) {
	color.Green("\tCOMANDO: " + op + "\n\tTIPO: " + mensaje)
}

func Confirmar(mensaje string) bool {
	color.Blue(mensaje + " (y/n)")
	var respuesta string
	fmt.Scanln(&respuesta)
	return Comparar(respuesta, "y")
}

func ArchivoExiste(ruta string) bool {
	if _, err := os.Stat(ruta); os.IsNotExist(err) {
		return false
	}
	return true
}

func EscribirBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)

	if err != nil {
		log.Fatal(err)
	}
}

func LeerDisco(path string) *Structs.MBR {
	m := Structs.MBR{}
	file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	defer file.Close()
	if err != nil {
		Error("FDISK", "Error al abrir el archivo")
		return nil
	}
	file.Seek(0, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &m)
	if err_ != nil {
		Error("FDSIK", "Error al leer el archivo")
		return nil
	}
	var mDir *Structs.MBR = &m
	return mDir
}

func leerBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number) //array de bytes

	_, err := file.Read(bytes) // Leido -> bytes
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func GetPartId(letra string, numero int) string {
	var st string = "65"
	nu := strconv.Itoa(numero)
	resultado := letra + nu + st
	return resultado
}

func GetPath(path string) []string {
	var result []string
	if path == "" {
		return result
	}
	aux := strings.Split(path, "/")
	for i := 1; i < len(aux); i++ {
		result = append(result, aux[i])
	}
	return result
}
