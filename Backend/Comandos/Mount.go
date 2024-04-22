package Comandos

import (
	"Proyecto/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var DiscMont [99]DiscoMontado

type DiscoMontado struct {
	Path        [150]byte
	Estado      byte
	Particiones [26]ParticionMontada
}

type ParticionMontada struct {
	Letra        byte
	Estado       byte
	Nombre       [20]byte
	Id_Particion [10]byte
}

var alfabeto = []byte{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}

func ValidarDatosMOUNT(context []string) {
	name := ""
	driveletter := ""
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "name") {
			name = comando[1]
		} else if Comparar(comando[0], "driveletter") {
			driveletter = strings.ReplaceAll(comando[1], "\"", "")
		}
	}
	if driveletter == "" || name == "" {
		Error("MOUNT", "El comando MOUNT requiere parámetros obligatorios")
		return
	}
	driveletter = driveletter + ".dsk"
	mount(driveletter, name)
	listaMount()
}

func mount(d string, n string) {
	file, error_ := os.Open(d)
	if error_ != nil {
		Error("MOUNT", "No se ha podido abrir el archivo.")
		return
	}

	disk := Structs.NewMBR()
	file.Seek(0, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &disk)
	if err_ != nil {
		Error("FDSIK", "Error al leer el archivo")
		return
	}
	file.Close()

	particion := BuscarParticiones(disk, n, d)
	if particion.Part_type == 'E' || particion.Part_type == 'L' {
		var nombre [16]byte
		copy(nombre[:], n)
		if particion.Part_name == nombre && particion.Part_type == 'E' {
			Error("MOUNT", "No se puede montar una partición extendida.")
			return
		} else {
			ebrs := GetLogicas(*particion, d)
			encontrada := false
			if len(ebrs) != 0 {
				for i := 0; i < len(ebrs); i++ {
					ebr := ebrs[i]
					nombreebr := ""
					for j := 0; j < len(ebr.Part_name); j++ {
						if ebr.Part_name[j] != 0 {
							nombreebr += string(ebr.Part_name[j])
						}
					}

					if Comparar(nombreebr, n) && ebr.Part_status == '1' {
						encontrada = true
						n = nombreebr
						break
					} else if nombreebr == n && ebr.Part_status == '0' {
						Error("MOUNT", "No se puede montar una partición Lógica eliminada.")
						return
					}
				}
				if !encontrada {
					Error("MOUNT", "No se encontró la partición Lógica.")
					return
				}
			}
		}
	}

	for i := 0; i < 99; i++ {
		var ruta [150]byte
		copy(ruta[:], d)
		if DiscMont[i].Path == ruta {
			for j := 0; j < 26; j++ {
				var nombre [20]byte
				copy(nombre[:], n)
				if DiscMont[i].Particiones[j].Nombre == nombre {
					Error("MOUNT", "Ya se ha montado la partición "+n)
					return
				}
				if DiscMont[i].Particiones[j].Estado == 0 {
					DiscMont[i].Particiones[j].Estado = 1
					DiscMont[i].Particiones[j].Letra = alfabeto[j]
					copy(DiscMont[i].Particiones[j].Nombre[:], n)
					IdTemp := GetPartId(string(alfabeto[i]), j+1)
					copy(DiscMont[i].Particiones[j].Id_Particion[:], IdTemp)

					Mensaje("MOUNT", "se ha realizado correctamente el mount -id= "+IdTemp)
					return
				}
			}
		}
	}
	for i := 0; i < 99; i++ {
		if DiscMont[i].Estado == 0 {
			copy(DiscMont[i].Path[:], d)
			DiscMont[i].Estado = 1

			for j := 0; j < 26; j++ {
				if DiscMont[i].Particiones[j].Estado == 0 {
					DiscMont[i].Particiones[j].Estado = 1
					DiscMont[i].Particiones[j].Letra = alfabeto[i]
					copy(DiscMont[i].Particiones[j].Nombre[:], n)
					IdTemp := GetPartId(string(alfabeto[i]), j+1)
					copy(DiscMont[i].Particiones[j].Id_Particion[:], IdTemp)

					Mensaje("MOUNT", "se ha realizado correctamente el mount -id= "+IdTemp)
					return
				}
			}
		}
	}
}

func ValidarDatosUNMOUNT(context []string) {
	id := ""
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comparar(comando[0], "id") {
			id = comando[1]
		}
	}
	if id == "" {
		Error("MOUNT", "El comando UNMOUNT requiere parámetros obligatorios")
		return
	}
	unmount(id)
}

func unmount(id string) {
	//haz un split de la posicion 0 y las ultimas 2 posiciones
	primeraLetra := id[0:1]
	Numero_C := id
	Numero_C = strings.Replace(Numero_C, primeraLetra, "", -1)
	Numero_C = strings.Replace(Numero_C, "65", "", -1)

	ultimosDos := id[len(id)-2:]
	if !(ultimosDos == "65") {
		Error("UNMOUNT", "El primer identificador no es válido.")
		return
	}
	i, _ := strconv.Atoi(Numero_C)
	for j := 0; j < 26; j++ {
		if DiscMont[i-1].Particiones[j].Estado == 1 {
			if DiscMont[i-1].Particiones[j].Letra == primeraLetra[0] {
				fmt.Println("Encontro la particion")
				path := ""
				for k := 0; k < len(DiscMont[i-1].Path); k++ {
					if DiscMont[i-1].Path[k] != 0 {
						path += string(DiscMont[i-1].Path[k])
					}
				}

				file, error := os.Open(strings.ReplaceAll(path, "\"", ""))
				if error != nil {
					Error("UNMOUNT", "No se ha encontrado el disco")
					return
				}
				disk := Structs.NewMBR()
				file.Seek(0, 0)

				data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
				buffer := bytes.NewBuffer(data)
				err_ := binary.Read(buffer, binary.BigEndian, &disk)

				if err_ != nil {
					Error("UNMOUNT", "Error al leer el archivo")
					return
				}
				file.Close()

				nombreParticion := ""
				for k := 0; k < len(DiscMont[i-1].Particiones[j].Nombre); k++ {
					if DiscMont[i-1].Particiones[j].Nombre[k] != 0 {
						nombreParticion += string(DiscMont[i-1].Particiones[j].Nombre[k])
					}
				}

				path = id[0:1]
				path = path + ".dsk"
				particion := GetMount("REP", id, &path)
				spr := Structs.NewSuperBloque()

				// Abrir el archivo una sola vez en modo lectura/escritura
				file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_RDWR, 0666)
				if err != nil {
					Error("REP", "No se ha encontrado el disco.")
					return
				}
				defer file.Close() // Asegurarse de cerrar el archivo al final

				file.Seek(particion.Part_start, 0)
				data = leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
				buffer = bytes.NewBuffer(data)
				err = binary.Read(buffer, binary.BigEndian, &spr)
				if err != nil {
					Error("REP", "Error al leer el archivo")
					return
				}

				fechaActual := time.Now().String()
				copy(spr.S_umtime[:], fechaActual)
				file.Seek(particion.Part_start, 0)
				// No es necesario buscar de nuevo, ya estamos en la posición correcta
				var binario2 bytes.Buffer
				binary.Write(&binario2, binary.BigEndian, spr)
				EscribirBytes(file, binario2.Bytes())
				//Formatea el discmount encontrado
				DiscMont[i-1].Particiones[j].Estado = 0
				DiscMont[i-1].Particiones[j].Letra = 0
				DiscMont[i-1].Particiones[j].Nombre = [20]byte{}
				DiscMont[i-1].Particiones[j].Id_Particion = [10]byte{}
				fmt.Println("Se ha desmontado la partición " + nombreParticion + " correctamente.")
				return

			}
		}
	}

}

func GetMount(comando string, id string, p *string) Structs.Particion {
	if !(id[len(id)-2] == '6' && id[len(id)-1] == '5') {
		Error(comando, "El primer identificador no es válido.")
		return Structs.Particion{}
	}
	letra := id[0:1]
	id = strings.ReplaceAll(id, "65", "")
	i, _ := strconv.Atoi(string(id[0] - 1))
	if i < 0 {
		Error(comando, "El primer identificador no es válido.")
		return Structs.Particion{}
	}
	for j := 0; j < 26; j++ {
		if DiscMont[i].Particiones[j].Estado == 1 {
			if DiscMont[i].Particiones[j].Letra == letra[0] {

				path := ""
				for k := 0; k < len(DiscMont[i].Path); k++ {
					if DiscMont[i].Path[k] != 0 {
						path += string(DiscMont[i].Path[k])
					}
				}

				file, error := os.Open(strings.ReplaceAll(path, "\"", ""))
				if error != nil {
					Error(comando, "No se ha encontrado el disco")
					return Structs.Particion{}
				}
				disk := Structs.NewMBR()
				file.Seek(0, 0)

				data := leerBytes(file, int(unsafe.Sizeof(Structs.MBR{})))
				buffer := bytes.NewBuffer(data)
				err_ := binary.Read(buffer, binary.BigEndian, &disk)

				if err_ != nil {
					Error("FDSIK", "Error al leer el archivo")
					return Structs.Particion{}
				}
				file.Close()

				nombreParticion := ""
				for k := 0; k < len(DiscMont[i].Particiones[j].Nombre); k++ {
					if DiscMont[i].Particiones[j].Nombre[k] != 0 {
						nombreParticion += string(DiscMont[i].Particiones[j].Nombre[k])
					}
				}
				*p = path
				return *BuscarParticiones(disk, nombreParticion, path)
			}
		}
	}
	return Structs.Particion{}
}

func listaMount() {
	fmt.Println("\n<-------------------------- LISTADO DE MOUNTS -------------------------->")
	for i := 0; i < 99; i++ {
		for j := 0; j < 26; j++ {
			if DiscMont[i].Particiones[j].Estado == 1 {
				nombre := ""
				Id_particion := ""
				for k := 0; k < len(DiscMont[i].Particiones[j].Nombre); k++ {
					if DiscMont[i].Particiones[j].Nombre[k] != 0 {
						nombre += string(DiscMont[i].Particiones[j].Nombre[k])
					}
				}
				for k := 0; k < len(DiscMont[i].Particiones[j].Id_Particion); k++ {
					if DiscMont[i].Particiones[j].Id_Particion[k] != 0 {
						Id_particion += string(DiscMont[i].Particiones[j].Id_Particion[k])
					}
				}
				fmt.Println("ID: " + Id_particion + " - Path: " + string(DiscMont[i].Path[:]) + " - Nombre: " + nombre + " - Letra: " + string(DiscMont[i].Particiones[j].Letra))

			}
		}
	}
}
