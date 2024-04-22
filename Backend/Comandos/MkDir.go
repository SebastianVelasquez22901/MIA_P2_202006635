package Comandos

import (
	"Proyecto/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"
)

func ValidarDatosDir(context []string) {
	path := ""
	r := ""

	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "path") {
			path = tk[1]
		} else if Comparar(tk[0], "r") {
			r = tk[0]
		}
	}
	if path == "" {
		Error("MKDir", "Se necesitan parametros obligatorio para crear un usuario.")
		return
	}
	MKDir(path, r)

}

func MKDir(p string, r string) {

	//Declaramos las variables
	/*

		blockCarpetasSize := int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))
		blockArchivosSize := int(unsafe.Sizeof(Structs.BloquesArchivos{}))
		inode := Structs.NewInodos()
		bloqueCarpeta := Structs.NewBloquesCarpetas()
		Content := Structs.NewContent()
		NuevoInodo := Structs.NewInodos()
	*/

	ruta := GetPath(p)
	fmt.Println("Ruta: ", ruta)

	var path string
	partition := GetMount("MKDIR", Logged.Id, &path)
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
	//PosRuta := 0
	CarpetaNueva := ruta[len(ruta)-1]
	fecha := time.Now().String()

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

	if len(ruta) == 1 {
		for i := 0; i < int(SB.S_inodes_count); i++ {
			Carpeta := true

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
				if inode.I_type == 1 {
					Carpeta = false
				} else if inode.I_type > 1 {
					fmt.Println("No valido")
				}
				for j := 0; j < len(inode.I_block); j++ {

					if inode.I_block[j] != -1 {
						if Carpeta {
							fmt.Println("Inodo Carpeta")

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
								Contenido := bc.B_content[k]
								//Nombre := strings.Trim(string(Contenido.B_name[:]), "\x00")
								if Contenido.B_inodo == -1 {

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
										copy(Journal.Operacion[:], "mkdir")
										copy(Journal.Path[:], p)
										copy(Journal.Fecha[:], fecha)
										copy(Journal.Contenido[:], "---")
										file.Seek(InicioJournal, 0)
										var binario bytes.Buffer
										binary.Write(&binario, binary.BigEndian, Journal)
										EscribirBytes(file, binario.Bytes())
									}

									inodetmp := Structs.NewInodos()
									inodetmp.I_uid = int64(Logged.Uid)
									inodetmp.I_gid = int64(Logged.Gid)
									inodetmp.I_size = int64(TamBloqueCarpeta)
									copy(inodetmp.I_atime[:], fecha)
									copy(inodetmp.I_ctime[:], fecha)
									copy(inodetmp.I_mtime[:], fecha)
									inodetmp.I_type = 0
									inodetmp.I_perm = 664
									SB.S_first_blo++
									inodetmp.I_block[0] = SB.S_first_blo
									SB.S_firts_ino++
									bc.B_content[k].B_inodo = SB.S_firts_ino
									Sumatoria := SB.S_inode_start + (int64(TamInodo) * SB.S_firts_ino)
									file.Seek(Sumatoria, 0)
									var tmpI bytes.Buffer
									binary.Write(&tmpI, binary.BigEndian, inodetmp)
									EscribirBytes(file, tmpI.Bytes())

									copy(bc.B_content[k].B_name[:], CarpetaNueva)

									file.Seek(PunteroBloquesCarpetas, 0)
									var tmpBc bytes.Buffer
									binary.Write(&tmpBc, binary.BigEndian, bc)
									EscribirBytes(file, tmpBc.Bytes())
									fb := Structs.NewBloquesCarpetas()
									copy(fb.B_content[0].B_name[:], ".")
									fb.B_content[0].B_inodo = 0
									copy(fb.B_content[1].B_name[:], "..")
									fb.B_content[1].B_inodo = 0
									copy(fb.B_content[2].B_name[:], "--")
									fb.B_content[2].B_inodo = -1
									copy(fb.B_content[3].B_name[:], "--")
									fb.B_content[3].B_inodo = -1
									PunteroBloquesCarpetas = SB.S_block_start + (int64(TamBloqueCarpeta) * int64(SB.S_first_blo-int64(CantidadBloquesArchivos)))

									file.Seek(PunteroBloquesCarpetas, 0)
									var bin5 bytes.Buffer
									binary.Write(&bin5, binary.BigEndian, fb)
									EscribirBytes(file, bin5.Bytes())

									file.Seek(partition.Part_start, 0)
									var binario333 bytes.Buffer
									binary.Write(&binario333, binary.BigEndian, SB)
									EscribirBytes(file, binario333.Bytes())
									return

								}

							}
							fmt.Println("No hay espacio en el bloque")
							PunteroTemporal := SB.S_block_start
							for {

								bc := Structs.NewBloquesCarpetas()
								file.Seek(PunteroTemporal, 0)
								data = leerBytes(file, TamBloqueCarpeta)
								buffer = bytes.NewBuffer(data)
								err_ = binary.Read(buffer, binary.BigEndian, &bc)
								if err_ != nil {
									Error("MkDir", "Error al leer el archivo")
									return
								}
								Contenido := bc.B_content[3]
								Nombre := strings.Trim(string(Contenido.B_name[:]), "\x00")
								fmt.Println("Nombre: ", Nombre)
								// Primera condición
								if strings.ContainsRune(Nombre, '�') {
									break
								} else {
									PunteroTemporal += int64(TamBloqueCarpeta)
								}
							}

							if inode.I_block[j+1] == -1 {

								fb := Structs.NewBloquesCarpetas()
								copy(fb.B_content[0].B_name[:], "--")
								fb.B_content[0].B_inodo = -1
								copy(fb.B_content[1].B_name[:], "--")
								fb.B_content[1].B_inodo = -1
								copy(fb.B_content[2].B_name[:], "--")
								fb.B_content[2].B_inodo = -1
								copy(fb.B_content[3].B_name[:], "--")
								fb.B_content[3].B_inodo = -1
								file.Seek(PunteroTemporal, 0)
								var bin5 bytes.Buffer
								binary.Write(&bin5, binary.BigEndian, fb)
								EscribirBytes(file, bin5.Bytes())

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
									copy(Journal.Operacion[:], "mkdir")
									copy(Journal.Path[:], p)
									copy(Journal.Fecha[:], fecha)
									copy(Journal.Contenido[:], "---")
									file.Seek(InicioJournal, 0)
									var binario bytes.Buffer
									binary.Write(&binario, binary.BigEndian, Journal)
									EscribirBytes(file, binario.Bytes())
								}

								SB.S_first_blo++
								inode.I_block[j+1] = SB.S_first_blo
								file.Seek(PunteroInodos, 0)
								var tmpI bytes.Buffer
								binary.Write(&tmpI, binary.BigEndian, inode)
								EscribirBytes(file, tmpI.Bytes())
								PunteroBloquesCarpetas = PunteroTemporal
							} else {
								Apunta := int(inode.I_block[j+1])
								PunteroBloquesCarpetas = SB.S_block_start + (int64(TamBloqueCarpeta) * int64(Apunta-CantidadBloquesArchivos))
							}

						}
					}
				}
				PunteroInodos += int64(TamInodo)
			}
		}
	} else {
		PunteroBloquesCarpetas = SB.S_block_start
		PosicionRuta := 0
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
			Contenido := bc.B_content[0]
			Nombre := strings.Trim(string(Contenido.B_name[:]), "\x00")
			if ruta[PosicionRuta] == CarpetaNueva {
				fmt.Println("Creando Carpeta")

				for k := 0; k < len(bc.B_content); k++ {
					if bc.B_content[k].B_inodo == -1 {

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
							copy(Journal.Operacion[:], "mkdir")
							copy(Journal.Path[:], p)
							copy(Journal.Fecha[:], fecha)
							copy(Journal.Contenido[:], "---")
							file.Seek(InicioJournal, 0)
							var binario bytes.Buffer
							binary.Write(&binario, binary.BigEndian, Journal)
							EscribirBytes(file, binario.Bytes())
						}

						inodetmp := Structs.NewInodos()
						inodetmp.I_uid = int64(Logged.Uid)
						inodetmp.I_gid = int64(Logged.Gid)
						inodetmp.I_size = int64(TamBloqueCarpeta)
						copy(inodetmp.I_atime[:], fecha)
						copy(inodetmp.I_ctime[:], fecha)
						copy(inodetmp.I_mtime[:], fecha)
						inodetmp.I_type = 0
						inodetmp.I_perm = 664
						SB.S_first_blo++
						inodetmp.I_block[0] = SB.S_first_blo
						SB.S_firts_ino++
						bc.B_content[k].B_inodo = SB.S_firts_ino
						Sumatoria := SB.S_inode_start + (int64(TamInodo) * SB.S_firts_ino)
						file.Seek(Sumatoria, 0)
						var tmpI bytes.Buffer
						binary.Write(&tmpI, binary.BigEndian, inodetmp)
						EscribirBytes(file, tmpI.Bytes())

						copy(bc.B_content[k].B_name[:], CarpetaNueva)

						file.Seek(PunteroBloquesCarpetas, 0)
						var tmpBc bytes.Buffer
						binary.Write(&tmpBc, binary.BigEndian, bc)
						EscribirBytes(file, tmpBc.Bytes())
						fb := Structs.NewBloquesCarpetas()
						copy(fb.B_content[0].B_name[:], ".")
						fb.B_content[0].B_inodo = 0
						copy(fb.B_content[1].B_name[:], "..")
						fb.B_content[1].B_inodo = 0
						copy(fb.B_content[2].B_name[:], "--")
						fb.B_content[2].B_inodo = -1
						copy(fb.B_content[3].B_name[:], "--")
						fb.B_content[3].B_inodo = -1
						PunteroBloquesCarpetas = SB.S_block_start + (int64(TamBloqueCarpeta) * int64(SB.S_first_blo-int64(CantidadBloquesArchivos)))

						file.Seek(PunteroBloquesCarpetas, 0)
						var bin5 bytes.Buffer
						binary.Write(&bin5, binary.BigEndian, fb)
						EscribirBytes(file, bin5.Bytes())

						file.Seek(partition.Part_start, 0)
						var binario333 bytes.Buffer
						binary.Write(&binario333, binary.BigEndian, SB)
						EscribirBytes(file, binario333.Bytes())
						return
					}
				}

				return
			} else if strings.ContainsRune(Nombre, '�') {
				break
			} else {
				PunteroBloquesCarpetas += int64(TamBloqueCarpeta)
			}

			for k := 0; k < len(bc.B_content); k++ {
				Contenido := bc.B_content[k]
				Nombre := strings.Trim(string(Contenido.B_name[:]), "\x00")
				ApuntadorInodo := Contenido.B_inodo
				if Nombre == ruta[PosicionRuta] {
					tempInode := Structs.NewInodos()
					PunteroInodos = SB.S_inode_start + (int64(TamInodo) * ApuntadorInodo)
					file.Seek(PunteroInodos, 0)
					data = leerBytes(file, TamInodo)
					buffer = bytes.NewBuffer(data)
					err_ = binary.Read(buffer, binary.BigEndian, &tempInode)
					if err_ != nil {
						Error("MkDir", "Error al leer el archivo")
						return
					}
					SiguienteBloque := tempInode.I_block[0]
					PunteroBloquesCarpetas = SB.S_block_start + (int64(TamBloqueCarpeta) * (SiguienteBloque - int64(CantidadBloquesArchivos)))
					PosicionRuta++

				}
			}

			// Primera condición

		}
	}

	//CantidadBloques := 0

}
