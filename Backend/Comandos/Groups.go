package Comandos

import (
	"Proyecto/Structs"
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func ValidarDatosGrupos(context []string, action string) {
	name := ""
	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "name") {
			name = tk[1]
		}
	}
	if name == "" {
		Error(action+"GRP", "No se encontró el parámetro name en el comando.")
		return
	}
	if Comparar(action, "MK") {
		mkgrp(name)
	} else if Comparar(action, "RM") {
		rmgrp(name)
	} else {
		Error(action+"GRP", "No se reconoce este comando.")
		return
	}
}

func mkgrp(n string) {
	if !Comparar(Logged.User, "root") {
		Error("MKGRP", "Solo el usuario \"root\" puede acceder a estos comandos.")
		return
	}

	var path string
	partition := GetMount("MKGRP", Logged.Id, &path)
	if string(partition.Part_status) == "0" {
		Error("MKGRP", "No se encontró la partición montada con el id: "+Logged.Id)
		return
	}
	//file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_RDWR, 0644)
	if err != nil {
		Error("MKGRP", "No se ha encontrado el disco.")
		return
	}

	super := Structs.NewSuperBloque()
	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &super)
	if err_ != nil {
		Error("MKGRP", "Error al leer el archivo")
		return
	}
	inode := Structs.NewInodos()
	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("MKGRP", "Error al leer el archivo")
		return
	}
	MitadBA := (partition.Part_size - super.S_block_start) / 2
	MitadBA = MitadBA + super.S_block_start
	PunteroBA := MitadBA
	TamBA := int64(unsafe.Sizeof(Structs.BloquesArchivos{}))
	var fb Structs.BloquesArchivos
	txt := ""
	for bloque := 1; bloque < 16; bloque++ {
		if inode.I_block[bloque-1] == -1 {
			break
		}
		file.Seek(PunteroBA, 0)

		data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &fb)

		if err_ != nil {
			Error("MKGRP", "Error al leer el archivo")
			return
		}

		for i := 0; i < len(fb.B_content); i++ {
			if fb.B_content[i] != 0 {
				txt += string(fb.B_content[i])
			}
		}
	}

	vctr := strings.Split(txt, "\n")
	c := 0
	for i := 0; i < len(vctr)-1; i++ {
		linea := vctr[i]
		if linea[2] == 'G' || linea[2] == 'g' {
			c++
			in := strings.Split(linea, ",")
			if in[2] == n {
				if linea[0] != '0' {
					Error("MKGRP", "EL nombre "+n+", ya está en uso.")
					return
				}
			}
		}
	}
	txt += strconv.Itoa(c+1) + ",G," + n + "\n"

	tam := len(txt)
	var cadenasS []string
	if tam > 64 {
		for tam > 64 {
			aux := ""
			for i := 0; i < 64; i++ {
				aux += string(txt[i])
			}
			cadenasS = append(cadenasS, aux)
			txt = strings.ReplaceAll(txt, aux, "")
			tam = len(txt)
		}
		if tam < 64 && tam != 0 {
			cadenasS = append(cadenasS, txt)
		}
	} else {
		cadenasS = append(cadenasS, txt)
	}
	if len(cadenasS) > 16 {
		Error("MKGRP", "Se ha llenado la cantidad de archivos posibles y no se pueden generar más.")
		return
	}
	file.Close()

	file, err = os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_RDWR, 0644)
	//file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("MKGRP", "No se ha encontrado el disco.")
		return
	}

	for i := 0; i < len(cadenasS); i++ {

		var fbAux Structs.BloquesArchivos
		if inode.I_block[i] == -1 {
			file.Seek(PunteroBA, 0)
			var binAux bytes.Buffer
			binary.Write(&binAux, binary.BigEndian, fbAux)
			EscribirBytes(file, binAux.Bytes())
			super.S_first_blo++
			inode.I_block[i] = super.S_first_blo
			file.Seek(partition.Part_start, 0)
			var binario333 bytes.Buffer
			binary.Write(&binario333, binary.BigEndian, super)
			EscribirBytes(file, binario333.Bytes())
			PunteroBA += TamBA

			TamJournal := int(unsafe.Sizeof(Structs.Journaling{}))
			InicioJournal := partition.Part_start + int64(unsafe.Sizeof(Structs.SuperBloque{}))
			Journal := Structs.NewJournal()
			fecha := time.Now().String()
			Ext3 := false
			for {

				file.Seek(InicioJournal, 0)
				data = leerBytes(file, TamJournal)
				buffer = bytes.NewBuffer(data)
				err_ = binary.Read(buffer, binary.BigEndian, &Journal)
				if err_ != nil {
					Error("MkDir", "Error al leer el archivo")
					return
				}
				if Journal.Journaling_start == 0 {
					break
				} else {
					InicioJournal += int64(TamJournal)
					Ext3 = true
				}
			}
			if Ext3 {
				Journal.Journaling_start = InicioJournal
				copy(Journal.Operacion[:], "mkfile")
				copy(Journal.Path[:], "/users.txt")
				copy(Journal.Fecha[:], fecha)
				copy(Journal.Contenido[:], cadenasS[i])
				file.Seek(InicioJournal, 0)
				var binario bytes.Buffer
				binary.Write(&binario, binary.BigEndian, Journal)
				EscribirBytes(file, binario.Bytes())
			}

		} else {
			fbAux = fb
		}

		copy(fbAux.B_content[:], cadenasS[i])

		file.Seek(PunteroBA, 0)
		var bin6 bytes.Buffer
		binary.Write(&bin6, binary.BigEndian, fbAux)
		EscribirBytes(file, bin6.Bytes())

	}

	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)
	var inodos bytes.Buffer
	binary.Write(&inodos, binary.BigEndian, inode)
	EscribirBytes(file, inodos.Bytes())

	Mensaje("MKGRP", "Grupo "+n+", creado correctamente!")

	file.Close()
}

func rmgrp(n string) {
	if !Comparar(Logged.User, "root") {
		Error("RMGRP", "Solo el usuario \"root\" puede acceder a estos comandos.")
		return
	}

	var path string
	partition := GetMount("RMGRP", Logged.Id, &path)
	if string(partition.Part_status) == "0" {
		Error("RMGRP", "No se encontró la partición montada con el id: "+Logged.Id)
		return
	}
	//file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("RMGRP", "No se ha encontrado el disco.")
		return
	}

	super := Structs.NewSuperBloque()
	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &super)
	if err_ != nil {
		Error("RMGRP", "Error al leer el archivo")
		return
	}
	inode := Structs.NewInodos()
	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("RMGRP", "Error al leer el archivo")
		return
	}

	var fb Structs.BloquesArchivos
	txt := ""
	for bloque := 1; bloque < 16; bloque++ {
		if inode.I_block[bloque-1] == -1 {
			break
		}
		file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(bloque-1), 0)

		data = leerBytes(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &fb)

		if err_ != nil {
			Error("RMGRP", "Error al leer el archivo")
			return
		}

		for i := 0; i < len(fb.B_content); i++ {
			if fb.B_content[i] != 0 {
				txt += string(fb.B_content[i])
			}
		}
	}

	aux := ""

	vctr := strings.Split(txt, "\n")
	existe := false
	for i := 0; i < len(vctr)-1; i++ {
		linea := vctr[i]
		if (linea[2] == 'G' || linea[2] == 'g') && linea[0] != '0' {
			in := strings.Split(linea, ",")
			if in[2] == n {
				existe = true
				aux += strconv.Itoa(0) + ",G," + in[2] + "\n"
				continue
			}
		}
		aux += linea + "\n"
	}
	if !existe {
		Error("RMGRP", "No se encontró el grupo \""+n+"\".")
		return
	}
	txt = aux

	tam := len(txt)
	var cadenasS []string
	if tam > 64 {
		for tam > 64 {
			aux := ""
			for i := 0; i < 64; i++ {
				aux += string(txt[i])
			}
			cadenasS = append(cadenasS, aux)
			txt = strings.ReplaceAll(txt, aux, "")
			tam = len(txt)
		}
		if tam < 64 && tam != 0 {
			cadenasS = append(cadenasS, txt)
		}
	} else {
		cadenasS = append(cadenasS, txt)
	}
	if len(cadenasS) > 16 {
		Error("RMGRP", "Se ha llenado la cantidad de archivos posibles y no se pueden generar más.")
		return
	}
	file.Close()

	file, err = os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	//file, err := os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Error("RMGRP", "No se ha encontrado el disco.")
		return
	}
	for i := 0; i < len(cadenasS); i++ {

		var fbAux Structs.BloquesArchivos
		if inode.I_block[i] == -1 {
			file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(i), 0)
			var binAux bytes.Buffer
			binary.Write(&binAux, binary.BigEndian, fbAux)
			EscribirBytes(file, binAux.Bytes())
		} else {
			fbAux = fb
		}

		copy(fbAux.B_content[:], cadenasS[i])

		file.Seek(super.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(i), 0)
		var bin6 bytes.Buffer
		binary.Write(&bin6, binary.BigEndian, fbAux)
		EscribirBytes(file, bin6.Bytes())

	}
	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)
	var inodos bytes.Buffer
	binary.Write(&inodos, binary.BigEndian, inode)
	EscribirBytes(file, inodos.Bytes())

	Mensaje("RMGRP", "Grupo "+n+", eliminado correctamente!")

	file.Close()
}
