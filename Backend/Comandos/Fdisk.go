package Comandos

import (
	"Proyecto/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

type Transition struct {
	partition int
	start     int
	end       int
	before    int
	after     int
}

var startValue int

func ValidarDatosFDISK(tokens []string) {
	if len(tokens) < 3 {
		Error("FDISK", "Se requieren al menos 3 parámetros para este comando.")
		return
	}
	size := ""
	driveletter := ""
	name := ""
	unit := ""
	tipo := ""
	fit := ""
	delete := ""
	add := ""

	error_ := false
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "fit") {
			if fit == "" {
				fit = tk[1]
			} else {
				Error("FDISK", "parametro fit repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "size") {
			if size == "" {
				size = tk[1]
			} else {
				Error("FDISK", "parametro SIZE repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "unit") {
			if unit == "" {
				unit = tk[1]
			} else {
				Error("FDISK", "parametro U repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "driveletter") {
			if driveletter == "" {
				driveletter = tk[1] + ".dsk"
			} else {
				Error("FDISK", "parametro driveletter repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "name") {
			if name == "" {
				name = tk[1]
			} else {
				Error("FDISK", "parametro name repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "type") {
			if tipo == "" {
				tipo = tk[1]
			} else {
				Error("FDISK", "parametro type repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "delete") {
			if delete == "" {
				delete = tk[1]
			} else {
				Error("FDISK", "parametro delete repetido en el comando: "+tk[0])
				return
			}
		} else if Comparar(tk[0], "add") {
			if add == "" {
				add = tk[1]
			} else {
				Error("FDISK", "parametro add repetido en el comando: "+tk[0])
				return
			}
		} else {
			Error("FDISK", "no se esperaba el parametro "+tk[0])
			error_ = true
			return
		}
	}
	if tipo == "" {
		tipo = "P"
	}
	if fit == "" {
		fit = "WF"
	}
	if unit == "" {
		unit = "K"
	}
	if error_ {
		return
	}
	if name == "" && driveletter == "" && size == "" {
		Error("FDISK", "name, driveletter y size son parámetros requeridos para este comando.")
		return
	} else if !Comparar(fit, "BF") && !Comparar(fit, "FF") && !Comparar(fit, "WF") {
		Error("FDISK", "valores en parametro fit no esperados")
		return
	} else if !Comparar(unit, "k") && !Comparar(unit, "m") && !Comparar(unit, "b") {
		Error("FDISK", "valores en parametro unit no esperados")
		return
	} else if !Comparar(tipo, "P") && !Comparar(tipo, "E") && !Comparar(tipo, "L") {
		fmt.Println(tipo)
		Error("FDISK", "valores en parametro type no esperados")
		return
	} else {
		if delete != "" {
			EliminarParticion(driveletter, name)
		} else if add != "" {

			AddParticion(driveletter, name, add, unit)

		} else {
			FDISK(size, driveletter, name, unit, tipo, fit, delete, add)
		}
	}
}

func AddParticion(d string, name string, valor string, u string) *Structs.Particion {
	mbr := LeerDisco(d)
	particiones := GetParticiones(*mbr)
	OP := ""
	for i := 0; i < len(particiones); i++ {
		particion := particiones[i]
		if particion.Part_status == "1"[0] {
			nombre := ""
			for j := 0; j < len(particion.Part_name); j++ {
				if particion.Part_name[j] != 0 {
					nombre += string(particion.Part_name[j])

				}

			}
			if Comparar(nombre, name) {
				num, _ := strconv.Atoi(valor)
				if Comparar(u, "b") || Comparar(u, "k") || Comparar(u, "m") {
					if Comparar(u, "k") {
						num = num * 1024
					} else if Comparar(u, "m") {
						num = num * 1024 * 1024
					}
				}
				if num > 0 {
					OP = "S"
				} else {
					OP = "R"
				}

				if i == 0 {
					if OP == "S" {
						fmt.Println(mbr.Mbr_tamano)
						fmt.Println(particiones[0].Part_size)
						if MenorQue(int(mbr.Mbr_tamano), Suma(int(particiones[0].Part_size), num)) {
							Error("FDISK", "No se puede aumentar la particion a un tamaño mayor que el disco")
							return nil
						}
						particiones[0].Part_size = int64(Suma(int(particiones[0].Part_size), num))
					} else {
						if particiones[0].Part_size < int64(num) {
							Error("FDISK", "No se puede reducir la particion a un tamaño menor que 0")
							return nil
						}
						particiones[0].Part_size = int64(Resta(int(particiones[0].Part_size), num))
					}
					mbr.Mbr_partition_1 = particiones[0]
				} else if i == 1 {
					if OP == "S" {
						if MenorQue(int(mbr.Mbr_tamano), Suma(int(particiones[1].Part_size), num)) {
							Error("FDISK", "No se puede aumentar la particion a un tamaño mayor que el disco")
							return nil
						}
						particiones[1].Part_size = int64(Suma(int(particiones[1].Part_size), num))
					} else {
						if particiones[1].Part_size < int64(num) {
							Error("FDISK", "No se puede reducir la particion a un tamaño menor que 0")
							return nil
						}
						particiones[1].Part_size = int64(Resta(int(particiones[1].Part_size), num))
					}

					mbr.Mbr_partition_2 = particiones[1]

				} else if i == 2 {

					if OP == "S" {
						if MenorQue(int(mbr.Mbr_tamano), Suma(int(particiones[2].Part_size), num)) {
							Error("FDISK", "No se puede aumentar la particion a un tamaño mayor que el disco")
							return nil
						}
						particiones[2].Part_size = int64(Suma(int(particiones[2].Part_size), num))
					} else {
						if particiones[2].Part_size < int64(num) {
							Error("FDISK", "No se puede reducir la particion a un tamaño menor que 0")
							return nil
						}
						particiones[2].Part_size = int64(Resta(int(particiones[2].Part_size), num))
					}

					mbr.Mbr_partition_3 = particiones[2]

				} else if i == 3 {
					if MenorQue(int(mbr.Mbr_tamano), Suma(int(particiones[3].Part_size), num)) {
						Error("FDISK", "No se puede aumentar la particion a un tamaño mayor que el disco")
						return nil
					}
					if OP == "S" {
						particiones[3].Part_size = int64(Suma(int(particiones[3].Part_size), num))
					} else {
						if particiones[3].Part_size < int64(num) {
							Error("FDISK", "No se puede reducir la particion a un tamaño menor que 0")
							return nil
						}
						particiones[3].Part_size = int64(Resta(int(particiones[3].Part_size), num))
					}

					mbr.Mbr_partition_4 = particiones[3]

				}

				// Elimina la particion

				file, err := os.OpenFile(strings.ReplaceAll(d, "\"", ""), os.O_WRONLY, os.ModeAppend)
				if err != nil {
					Error("FDISK", "Error al abrir el archivo")
					return nil
				}
				file.Seek(0, 0)
				var binario2 bytes.Buffer
				binary.Write(&binario2, binary.BigEndian, mbr)
				EscribirBytes(file, binario2.Bytes())
				return &particion
			}

		}
	}
	Error("FDISK", "No se encontro la particion")

	return nil
}

func FDISK(s string, d string, n string, u string,
	t string, f string, delete string, add string) {

	//size, driveletter, name, unit, tipo, fit, delete, add

	startValue = 0
	i, error_ := strconv.Atoi(s)
	if error_ != nil {
		Error("FDISK", "Size debe ser un número entero")
		return
	}
	if i <= 0 {
		Error("FDISK", "Size debe ser mayor que 0")
		return
	}
	if Comparar(u, "b") || Comparar(u, "k") || Comparar(u, "m") {
		if Comparar(u, "k") {
			i = i * 1024
		} else if Comparar(u, "m") {
			i = i * 1024 * 1024
		}
	} else {
		Error("FDISK", "Unit no contiene los valores esperados.")
		return
	}
	if !(Comparar(t, "p") || Comparar(t, "e") || Comparar(t, "l")) {
		Error("FDISK", "Type no contiene los valores esperados.")
		return
	}
	if !(Comparar(f, "bf") || Comparar(f, "ff") || Comparar(f, "wf")) {
		Error("FDISK", "Fit no contiene los valores esperados.")
		return
	}
	mbr := LeerDisco(d)
	particiones := GetParticiones(*mbr)
	if MayorQue(i, int(mbr.Mbr_tamano)) {
		Error("FDISK", "El tamaño de la partición es mayor que el tamaño del disco.")
		return

	}

	var between []Transition

	usado := 0
	ext := 0
	c := 0
	base := int(unsafe.Sizeof(Structs.MBR{}))
	extended := Structs.NewParticion()

	for j := 0; j < len(particiones); j++ {
		prttn := particiones[j]
		if prttn.Part_status == '1' {
			var trn Transition
			trn.partition = c
			trn.start = int(prttn.Part_start)
			trn.end = int(prttn.Part_start + prttn.Part_size)
			trn.before = trn.start - base
			base = trn.end
			if usado != 0 {
				between[usado-1].after = trn.start - (between[usado-1].end)
			}
			between = append(between, trn)
			usado++

			if prttn.Part_type == "e"[0] || prttn.Part_type == "E"[0] {
				ext++
				extended = prttn
			}
		}
		if usado == 4 && !Comparar(t, "l") {
			Error("FDISK", "Limite de particiones alcanzado")
			return
		} else if ext == 1 && Comparar(t, "e") {
			Error("FDISK", "Solo se puede crear una partición extendida")
			return
		}
		c++
	}
	if ext == 0 && Comparar(t, "l") {
		Error("FDISK", "Aún no se han creado particiones extendidas, no se puede agregar una lógica.")
		return
	}
	if usado != 0 {
		between[len(between)-1].after = int(mbr.Mbr_tamano) - between[len(between)-1].end
	}
	regresa := BuscarParticiones(*mbr, n, d)
	if regresa != nil {
		Error("FDISK", "El nombre: "+n+", ya está en uso.")
		return
	}
	temporal := Structs.NewParticion()
	temporal.Part_status = '1'
	temporal.Part_size = int64(i)
	temporal.Part_type = strings.ToUpper(t)[0]
	temporal.Part_fit = strings.ToUpper(f)[0]
	copy(temporal.Part_name[:], n)

	if Comparar(t, "l") {
		Logica(temporal, extended, d)
		return
	}
	mbr = AjusteFit(*mbr, temporal, between, particiones, usado)
	if mbr == nil {
		return
	}
	file, err := os.OpenFile(strings.ReplaceAll(d, "\"", ""), os.O_WRONLY, os.ModeAppend)
	if err != nil {
		Error("FDISK", "Error al abrir el archivo")
		return
	}
	file.Seek(0, 0)
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, mbr)
	EscribirBytes(file, binario2.Bytes())
	if Comparar(t, "E") {
		ebr := Structs.NewEBR()
		ebr.Part_status = '0'
		ebr.Part_start = int64(startValue)
		ebr.Part_size = 0
		ebr.Part_next = -1

		file.Seek(int64(startValue), 0) //5200
		var binario3 bytes.Buffer
		binary.Write(&binario3, binary.BigEndian, ebr)
		EscribirBytes(file, binario3.Bytes())
		Mensaje("FDISK", "Partición Extendida: "+n+", creada correctamente.")
		return
	}
	file.Close()
	Mensaje("FDISK", "Partición Primaria: "+n+", creada correctamente.")
}

func EliminarParticion(d string, name string) *Structs.Particion {
	mbr := LeerDisco(d)
	particiones := GetParticiones(*mbr)

	for i := 0; i < len(particiones); i++ {
		particion := particiones[i]
		if particion.Part_status == "1"[0] {
			nombre := ""
			for j := 0; j < len(particion.Part_name); j++ {
				if particion.Part_name[j] != 0 {
					nombre += string(particion.Part_name[j])

				}

			}
			if Comparar(nombre, name) {
				fmt.Println(i)

				if i == 0 {

					particiones[0].Part_status = '0'
					particiones[0].Part_type = 'P'
					particiones[0].Part_fit = 'F'
					particiones[0].Part_start = -1
					particiones[0].Part_size = 0
					particiones[0].Part_name = [16]byte{}

					mbr.Mbr_partition_1 = particiones[0]

				} else if i == 1 {
					particiones[1].Part_status = '0'
					particiones[1].Part_type = 'P'
					particiones[1].Part_fit = 'F'
					particiones[1].Part_start = -1
					particiones[1].Part_size = 0
					particiones[1].Part_name = [16]byte{}

					mbr.Mbr_partition_2 = particiones[1]

				} else if i == 2 {

					particiones[2].Part_status = '0'
					particiones[2].Part_type = 'P'
					particiones[2].Part_fit = 'F'
					particiones[2].Part_start = -1
					particiones[2].Part_size = 0
					particiones[2].Part_name = [16]byte{}

					mbr.Mbr_partition_3 = particiones[2]

				} else if i == 3 {

					particiones[3].Part_status = '0'
					particiones[3].Part_type = 'P'
					particiones[3].Part_fit = 'F'
					particiones[3].Part_start = -1
					particiones[3].Part_size = 0
					particiones[3].Part_name = [16]byte{}

					mbr.Mbr_partition_4 = particiones[3]

				}

				// Elimina la particion

				file, err := os.OpenFile(strings.ReplaceAll(d, "\"", ""), os.O_WRONLY, os.ModeAppend)
				if err != nil {
					Error("FDISK", "Error al abrir el archivo")
					return nil
				}
				file.Seek(0, 0)
				var binario2 bytes.Buffer
				binary.Write(&binario2, binary.BigEndian, mbr)
				EscribirBytes(file, binario2.Bytes())
				return &particion
			}

		}
	}
	fmt.Println("No encontro la particion")

	return nil
}

func GetParticiones(disco Structs.MBR) []Structs.Particion {
	var v []Structs.Particion
	v = append(v, disco.Mbr_partition_1)
	v = append(v, disco.Mbr_partition_2)
	v = append(v, disco.Mbr_partition_3)
	v = append(v, disco.Mbr_partition_4)
	return v
}

func BuscarParticiones(mbr Structs.MBR, name string, path string) *Structs.Particion {
	var particiones [4]Structs.Particion
	particiones[0] = mbr.Mbr_partition_1
	particiones[1] = mbr.Mbr_partition_2
	particiones[2] = mbr.Mbr_partition_3
	particiones[3] = mbr.Mbr_partition_4

	ext := false
	extended := Structs.NewParticion()
	for i := 0; i < len(particiones); i++ {
		particion := particiones[i]
		if particion.Part_status == "1"[0] {
			nombre := ""
			for j := 0; j < len(particion.Part_name); j++ {
				if particion.Part_name[j] != 0 {
					nombre += string(particion.Part_name[j])
				}
			}
			if Comparar(nombre, name) {
				return &particion
			} else if particion.Part_type == "E"[0] || particion.Part_type == "e"[0] {
				ext = true
				extended = particion
			}
		}
	}

	if ext {
		ebrs := GetLogicas(extended, path)
		for i := 0; i < len(ebrs); i++ {
			ebr := ebrs[i]
			if ebr.Part_status == '1' {
				nombre := ""
				for j := 0; j < len(ebr.Part_name); j++ {
					if ebr.Part_name[j] != 0 {
						nombre += string(ebr.Part_name[j])
					}
				}
				fmt.Println(nombre)
				if Comparar(nombre, name) {
					tmp := Structs.NewParticion()
					tmp.Part_status = '1'
					tmp.Part_type = 'L'
					tmp.Part_fit = ebr.Part_fit
					tmp.Part_start = ebr.Part_start
					tmp.Part_size = ebr.Part_size
					copy(tmp.Part_name[:], ebr.Part_name[:])
					return &tmp
				}
			}
		}
	}
	return nil
}

func GetLogicas(particion Structs.Particion, path string) []Structs.EBR {
	var ebrs []Structs.EBR
	file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("FDISK", "Error al abrir el archivo")
		return nil
	}
	file.Seek(0, 0)
	tmp := Structs.NewEBR()
	file.Seek(particion.Part_start, 0)

	data := leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &tmp)
	if err_ != nil {
		Error("FDSIK", "Error al leer el archivo")
		return nil
	}
	for {
		if int(tmp.Part_next) != -1 && int(tmp.Part_status) != 0 {
			ebrs = append(ebrs, tmp)
			file.Seek(tmp.Part_next, 0)

			data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
			buffer = bytes.NewBuffer(data)
			err_ = binary.Read(buffer, binary.BigEndian, &tmp)
			if err_ != nil {
				Error("FDSIK", "Error al leer el archivo")
				return nil
			}
		} else {
			file.Close()
			break
		}
	}

	return ebrs
}

func Logica(particion Structs.Particion, ep Structs.Particion, path string) {
	logic := Structs.NewEBR()
	logic.Part_status = '1'
	logic.Part_fit = particion.Part_fit
	logic.Part_size = particion.Part_size
	logic.Part_next = -1
	copy(logic.Part_name[:], particion.Part_name[:])

	file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("FDISK", "Error al abrir el archivo del disco.")
		return
	}
	file.Seek(0, 0)

	tmp := Structs.NewEBR()
	tmp.Part_status = 0
	tmp.Part_size = 0
	tmp.Part_next = -1
	file.Seek(ep.Part_start, 0) //0

	data := leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &tmp)

	if err_ != nil {
		Error("FDSIK", "Error al leer el archivo")
		return
	}
	if err != nil {
		Error("FDISK", "Error al abrir el archivo del disco.")
		return
	}
	var size int64 = 0
	file.Close()
	for {
		size += int64(unsafe.Sizeof(Structs.EBR{})) + tmp.Part_size
		if (tmp.Part_size == 0 && tmp.Part_next == -1) || (tmp.Part_size == 0 && tmp.Part_next == 0) {
			file, err = os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
			logic.Part_start = tmp.Part_start
			logic.Part_next = logic.Part_start + logic.Part_size + int64(unsafe.Sizeof(Structs.EBR{}))
			if (ep.Part_size - size) <= logic.Part_size {
				Error("FDISK", "No queda más espacio para crear más particiones lógicas")
				return
			}
			file.Seek(logic.Part_start, 0)

			var binario2 bytes.Buffer
			binary.Write(&binario2, binary.BigEndian, logic)
			EscribirBytes(file, binario2.Bytes())
			nombre := ""
			for j := 0; j < len(particion.Part_name); j++ {
				nombre += string(particion.Part_name[j])
			}
			file.Seek(logic.Part_next, 0)
			addLogic := Structs.NewEBR()
			addLogic.Part_status = '0'
			addLogic.Part_next = -1
			addLogic.Part_start = logic.Part_next

			file.Seek(addLogic.Part_start, 0)

			var binarioLogico bytes.Buffer
			binary.Write(&binarioLogico, binary.BigEndian, addLogic)
			EscribirBytes(file, binarioLogico.Bytes())

			Mensaje("FDISK", "Partición Lógica: "+nombre+", creada correctamente.")
			file.Close()
			return
		}
		file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
		if err != nil {
			Error("FDISK", "Error al abrir el archivo del disco.")
			return
		}
		file.Seek(tmp.Part_next, 0)
		data = leerBytes(file, int(unsafe.Sizeof(Structs.EBR{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &tmp)

		if err_ != nil {
			Error("FDSIK", "Error al leer el archivo")
			return
		}
	}
}

func AjusteFit(mbr Structs.MBR, p Structs.Particion, t []Transition, ps []Structs.Particion, u int) *Structs.MBR {
	if u == 0 {
		p.Part_start = int64(unsafe.Sizeof(mbr))
		startValue = int(p.Part_start)
		mbr.Mbr_partition_1 = p
		return &mbr
	} else {

		// un for que me obtenga el fit que voy a usar
		fit := ""
		for _, b := range mbr.Dsk_fit {
			if b != 0 {
				fit += string(b)
			} else {
				break
			}
		}

		var usar Transition
		c := 0
		for i := 0; i < len(t); i++ {
			tr := t[i]
			if c == 0 {
				usar = tr
				c++
				continue
			}

			if Comparar(fit, "FF") {
				if int64(usar.before) >= p.Part_size || int64(usar.after) >= p.Part_size {
					break
				}
				usar = tr
			} else if Comparar(fit, "BF") {
				if int64(tr.before) >= p.Part_size || int64(usar.after) < p.Part_size {
					usar = tr
				} else {
					if int64(tr.before) >= p.Part_size || int64(tr.after) >= p.Part_size {
						b1 := usar.before - int(p.Part_size)
						a1 := usar.after - int(p.Part_size)
						b2 := tr.before - int(p.Part_size)
						a2 := tr.after - int(p.Part_size)

						if (b1 < b2 && b1 < a2) || (a1 < b2 && a1 < a2) {
							c++
							continue
						}
						usar = tr
					}
				}
			} else if Comparar(fit, "WF") {
				if int64(usar.before) >= p.Part_size || int64(usar.after) < p.Part_size {
					usar = tr
				} else {
					if int64(tr.before) >= p.Part_size || int64(tr.after) >= p.Part_size {
						b1 := usar.before - int(p.Part_size)
						a1 := usar.after - int(p.Part_size)
						b2 := tr.before - int(p.Part_size)
						a2 := tr.after - int(p.Part_size)

						if (b1 > b2 && b1 > a2) || (a1 > b2 && a1 > a2) {
							c++
							continue
						}
						usar = tr
					}
				}
			}
			c++
		}
		if usar.before >= int(p.Part_size) || usar.after >= int(p.Part_size) {
			if Comparar(fit, "FF") {
				if usar.before >= int(p.Part_size) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			} else if Comparar(fit, "BF") {
				b1 := usar.before - int(p.Part_size)
				a1 := usar.after - int(p.Part_size)

				if (usar.before >= int(p.Part_size) && b1 < a1) || usar.after < int(p.Part_start) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			} else if Comparar(fit, "WF") {
				b1 := usar.before - int(p.Part_size)
				a1 := usar.after - int(p.Part_size)

				if (usar.before >= int(p.Part_size) && b1 > a1) || usar.after < int(p.Part_start) {
					p.Part_start = int64(usar.start - usar.before)
					startValue = int(p.Part_start)
				} else {
					p.Part_start = int64(usar.end)
					startValue = int(p.Part_start)
				}
			}
			var partitions [4]Structs.Particion
			for i := 0; i < len(ps); i++ {
				partitions[i] = ps[i]
			}

			for i := 0; i < len(partitions); i++ {
				partition := partitions[i]
				if partition.Part_status != '1' {
					partitions[i] = p
					break
				}
			}
			mbr.Mbr_partition_1 = partitions[0]
			mbr.Mbr_partition_2 = partitions[1]
			mbr.Mbr_partition_3 = partitions[2]
			mbr.Mbr_partition_4 = partitions[3]
			return &mbr
		} else {
			Error("FDISK", "No hay espacio suficiente.")
			return nil
		}
	}
}
