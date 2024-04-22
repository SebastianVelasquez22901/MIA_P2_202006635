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

func ValidarDatosChown(context []string) {
	path := ""
	user := ""
	r := ""

	for i := 0; i < len(context); i++ {
		token := context[i]
		tk := strings.Split(token, "=")
		if Comparar(tk[0], "path") {
			path = tk[1]
		} else if Comparar(tk[0], "r") {
			r = tk[0]
		} else if Comparar(tk[0], "user") {
			user = tk[1]
		}
	}
	if path == "" {
		Error("MKDir", "Se necesitan parametros obligatorio para crear un usuario.")
		return
	}
	Chown(path, r, user)

}

func Chown(p string, r string, u string) {
	var path string
	partition := GetMount("CHOWN", Logged.Id, &path)
	if string(partition.Part_status) == "0" {
		Error("LOGIN", "No se encontró la partición montada con el id: "+Logged.Id)
		return
	}
	//file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_RDWR, 0644)
	if err != nil {
		Error("CHOWN", "No se ha encontrado el disco.")
		return
	}

	super := Structs.NewSuperBloque()
	file.Seek(partition.Part_start, 0)
	data := leerBytes(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &super)
	if err_ != nil {
		Error("LOGIN", "Error al leer el archivo")
		return
	}
	inode := Structs.NewInodos()
	file.Seek(super.S_inode_start+int64(unsafe.Sizeof(Structs.Inodos{})), 0)
	data = leerBytes(file, int(unsafe.Sizeof(Structs.Inodos{})))
	buffer = bytes.NewBuffer(data)
	err_ = binary.Read(buffer, binary.BigEndian, &inode)
	if err_ != nil {
		Error("LOGIN", "Error al leer el archivo")
		return
	}
	var fb Structs.BloquesArchivos
	txt := ""
	MitadBA := (partition.Part_size - super.S_block_start) / 2
	MitadBA = MitadBA + super.S_block_start
	TamBA := int64(unsafe.Sizeof(Structs.BloquesArchivos{}))
	PunteroBA := MitadBA
	for bloque := 1; bloque < 16; bloque++ {
		if inode.I_block[bloque-1] == -1 {
			break
		}
		file.Seek(PunteroBA, 0)
		data = leerBytes(file, int(TamBA))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &fb)

		if err_ != nil {
			Error("LOGIN", "Error al leer el archivo")
			return
		}
		PunteroBA += TamBA

		for i := 0; i < len(fb.B_content); i++ {
			if fb.B_content[i] != 0 {
				txt += string(fb.B_content[i])
			}
		}
	}

	vctr := strings.Split(txt, "\n")
	for i := 0; i < len(vctr)-1; i++ {
		linea := vctr[i]
		if linea[2] == 'U' || linea[2] == 'u' {
			in := strings.Split(linea, ",")
			if Comparar(in[3], u) && in[0] != "0" {
				idGrupo := "0"
				existe := false
				for j := 0; j < len(vctr)-1; j++ {
					line := vctr[j]
					if (line[2] == 'G' || line[2] == 'g') && line[0] != '0' {
						inG := strings.Split(line, ",")
						if inG[2] == in[2] {
							idGrupo = inG[0]
							existe = true
							break
						}
					}
				}
				if !existe {
					Error("Login", "No se encontró el grupo \""+in[2]+"\".")
					return
				}

				ruta := GetPath(p)

				PunteroBloquesCarpetas := super.S_block_start
				TamBloqueCarpeta := int(unsafe.Sizeof(Structs.BloquesCarpetas{}))
				TamInodo := int(unsafe.Sizeof(Structs.Inodos{}))
				PunteroInodos := super.S_inode_start
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

					for k := 0; k < len(bc.B_content); k++ {
						Contenido := bc.B_content[k]
						Nombre := strings.Trim(string(Contenido.B_name[:]), "\x00")
						ApuntadorInodo := Contenido.B_inodo
						if Nombre == ruta[len(ruta)-1] {
							tempInode := Structs.NewInodos()
							PunteroInodos = super.S_inode_start + (int64(TamInodo) * (ApuntadorInodo))
							file.Seek(PunteroInodos, 0)
							data = leerBytes(file, TamInodo)
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &tempInode)
							if err_ != nil {
								Error("MkDir", "Error al leer el archivo")
								return
							}
							Uid, err := strconv.ParseInt(in[0], 10, 64)
							if err != nil {
								Error("CHOWN", "Error al convertir el Uid")
								return
							}
							Gid, err := strconv.ParseInt(idGrupo, 10, 64)
							if err != nil {
								Error("CHOWN", "Error al convertir el Uid")
								return
							}
							tempInode.I_uid = Uid
							tempInode.I_gid = Gid
							file.Seek(PunteroInodos, 0)
							var tmpI bytes.Buffer
							binary.Write(&tmpI, binary.BigEndian, tempInode)
							EscribirBytes(file, tmpI.Bytes())
							return
						}
					}
					if strings.ContainsRune(Nombre, '�') {
						break
					} else {
						PunteroBloquesCarpetas += int64(TamBloqueCarpeta)
					}
				}

				//Uid := strconv.Atoi(in[0])
				//Gid := strconv.Atoi(idGrupo)
				fmt.Println("Usuario: " + in[3] + " Grupo: " + in[2] + " Uid: " + in[0] + " Gid: " + idGrupo)

				return
			}
		}
	}
	Error("LOGIN", "No se encontró el usuario "+u+" o la contraseña es incorrecta.")
	//CantidadBloques := 0

}
