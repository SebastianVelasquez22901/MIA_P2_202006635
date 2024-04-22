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

func ValidarDatosFile(context []string) {
	path := ""
	r := ""
	size := "0"
	count := ""

	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "path") {
			path = tk[1]
		} else if Comparar(tk[0], "r") {
			r = tk[0]
		} else if Comparar(tk[0], "size") {
			size = tk[1]
		} else if Comparar(tk[0], "count") {
			count = tk[1]
		}
	}
	if path == "" {
		Error("MKFILE", "Se necesitan parametros obligatorio para crear un usuario.")
		return
	}
	MKFILE(path, r, size, count)

}

func MKFILE(p string, r string, size string, count string) {

	//Declaramos las variables
	/*
		inodoSize := int64(unsafe.Sizeof(Structs.Inodos{}))
		blockCarpetasSize := int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))
		blockArchivosSize := int(unsafe.Sizeof(Structs.BloquesArchivos{}))
		inode := Structs.NewInodos()
		bloqueCarpeta := Structs.NewBloquesCarpetas()
		Content := Structs.NewContent()
		NuevoInodo := Structs.NewInodos()
	*/
	valTAm, roo := strconv.ParseInt(size, 10, 64)
	if roo != nil {
		Error("MKFILE", "Error al leer el tamaño")
		return
	}

	if valTAm < 0 {
		Error("MKFILE", "El tamaño no puede ser negativo")
		return
	}

	ruta := GetPath(p)
	fmt.Println("Ruta: ", ruta)

	var path string
	partition := GetMount("MKUSR", Logged.Id, &path)
	if string(partition.Part_status) == "0" {
		Error("MKUSR", "No se encontró la partición montada con el id: "+Logged.Id)
		return
	}
	SB := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_RDWR, 0644)
	if err != nil {
		Error("MKGRP", "No se ha encontrado el disco.")
		return
	}
	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &SB)
	if err_ != nil {
		Error("MKGRP", "Error al leer el archivo")
		return
	}
	MitadBA := (partition.Part_size - SB.S_block_start) / 2
	MitadBA = MitadBA + SB.S_block_start
	inode := Structs.NewInodos()
	PunteroInodos := SB.S_inode_start
	PunteroBloquesCarpetas := SB.S_block_start
	PunteroBloquesArchivos := MitadBA
	TamBloqueCarpeta := int(unsafe.Sizeof(Structs.BloquesCarpetas{}))
	TamInodo := int(unsafe.Sizeof(Structs.Inodos{}))
	TamBloqueArchivo := int(unsafe.Sizeof(Structs.BloquesArchivos{}))
	fecha := time.Now().String()
	//PosRuta := 0
	ArchivoNuevo := ruta[len(ruta)-1]
	CantidadBloquesArchivos := 0
	for {
		var fb Structs.BloquesArchivos
		file.Seek(PunteroBloquesArchivos, 0)
		data = leerBytes(file, TamBloqueArchivo)
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &fb)
		if err_ != nil {
			Error("REP", "Error al leer el archivo")
			return
		}
		Contenido := fb.B_content[0]
		if Contenido != 255 {
			CantidadBloquesArchivos++
		} else {
			break
		}
		PunteroBloquesArchivos += int64(TamBloqueArchivo)
	}

	CantidadBloquesCarpetas := 0
	for {
		bc := Structs.NewBloquesCarpetas()
		file.Seek(PunteroBloquesCarpetas, 0)
		data = leerBytes(file, TamBloqueCarpeta)
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &bc)
		if err_ != nil {
			Error("MkDir", "Error al leer el archivo")
			return
		}
		Contenido := bc.B_content[1]
		Nombre := strings.Trim(string(Contenido.B_name[:]), "\x00")
		// Primera condición
		if strings.ContainsRune(Nombre, '�') {
			break
		} else {
			CantidadBloquesCarpetas++
			PunteroBloquesCarpetas += int64(TamBloqueCarpeta)
		}

	}
	PunteroBloquesCarpetas = SB.S_block_start

	if len(ruta) == 1 {
		for i := 0; i < int(SB.S_inodes_count); i++ {

			file.Seek(PunteroInodos, 0)
			data = leerBytes(file, TamInodo)
			buffer = bytes.NewBuffer(data)
			err_ = binary.Read(buffer, binary.BigEndian, &inode)
			if err_ != nil {
				Error("MkDir", "Error al leer el archivo")
				return
			}

			if inode.I_size != -1 {

				//Carpeta 0 Archivo 1

				for j := 0; j < len(inode.I_block); j++ {

					if inode.I_block[j] != -1 {
						Apunta := inode.I_block[j]
						if Apunta == 0 {
							PunteroBloquesCarpetas = SB.S_block_start + (int64(Apunta) * int64(TamBloqueCarpeta))
						} else {
							resut := Apunta - 2
							PunteroBloquesCarpetas = SB.S_block_start + resut*int64(TamBloqueCarpeta)
						}
						bc := Structs.NewBloquesCarpetas()
						file.Seek(PunteroBloquesCarpetas, 0)
						data = leerBytes(file, TamBloqueCarpeta)
						buffer = bytes.NewBuffer(data)
						err_ = binary.Read(buffer, binary.BigEndian, &bc)
						if err_ != nil {
							Error("MkDir", "Error al leer el archivo")
							return
						}
						for k := 0; k < len(bc.B_content); k++ {
							if bc.B_content[k].B_inodo == -1 {

								tam, err := strconv.ParseInt(size, 10, 64)
								if err != nil {
									return
								}
								fmt.Println("Espacio disponible")
								SB.S_firts_ino++
								inodetmp := Structs.NewInodos()
								inodetmp.I_uid = int64(Logged.Uid)
								inodetmp.I_gid = int64(Logged.Gid)
								inodetmp.I_size = tam
								copy(inodetmp.I_atime[:], fecha)
								copy(inodetmp.I_ctime[:], fecha)
								copy(inodetmp.I_mtime[:], fecha)
								inodetmp.I_type = 1
								inodetmp.I_perm = 664
								SB.S_first_blo++
								inodetmp.I_block[0] = SB.S_first_blo

								Sumatoria := SB.S_inode_start + (int64(TamInodo) * SB.S_firts_ino)
								file.Seek(Sumatoria, 0)
								var tmpI bytes.Buffer
								binary.Write(&tmpI, binary.BigEndian, inodetmp)
								EscribirBytes(file, tmpI.Bytes())

								bc.B_content[k].B_inodo = SB.S_firts_ino
								copy(bc.B_content[k].B_name[:], ArchivoNuevo)
								file.Seek(PunteroBloquesCarpetas, 0)
								var tmpBc bytes.Buffer
								binary.Write(&tmpBc, binary.BigEndian, bc)
								EscribirBytes(file, tmpBc.Bytes())

								resultado := ""

								for i := 1; i <= int(tam); i++ {
									resultado += strconv.Itoa(i % 10)
								}

								TamJournal := int(unsafe.Sizeof(Structs.Journaling{}))
								InicioJournal := partition.Part_start + int64(unsafe.Sizeof(Structs.SuperBloque{}))
								Journal := Structs.NewJournal()

								Ext3 := false

								for {

									file.Seek(InicioJournal, 0)
									data = leerBytes(file, TamJournal)
									buffer = bytes.NewBuffer(data)
									err_ = binary.Read(buffer, binary.BigEndian, &Journal)
									if err_ != nil {
										Error("Mkfile", "Error al leer el archivo")
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
									copy(Journal.Path[:], p)
									copy(Journal.Fecha[:], fecha)
									copy(Journal.Contenido[:], resultado)
									file.Seek(InicioJournal, 0)
									var binario bytes.Buffer
									binary.Write(&binario, binary.BigEndian, Journal)
									EscribirBytes(file, binario.Bytes())
								}

								var fileb Structs.BloquesArchivos
								copy(fileb.B_content[:], resultado)
								PunteroBloquesArchivos = MitadBA + (int64(CantidadBloquesArchivos) * int64(TamBloqueArchivo))
								file.Seek(PunteroBloquesArchivos, 0)
								var bin6 bytes.Buffer
								binary.Write(&bin6, binary.BigEndian, fileb)
								EscribirBytes(file, bin6.Bytes())

								file.Seek(partition.Part_start, 0)
								var binario333 bytes.Buffer
								binary.Write(&binario333, binary.BigEndian, SB)
								EscribirBytes(file, binario333.Bytes())

								return

							}
						}

					}
				}
			}
			PunteroInodos += int64(TamInodo)
		}
	}

}
