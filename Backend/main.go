package main

import (
	"Proyecto/Comandos"
	"Proyecto/Reportes"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// exec -path=/home/daniel/Escritorio/MIA_1S2024/Ejemplos_Proyecto/calificacion.script

var logued = false

type DatosEntrada struct {
	Comandos []string `json:"comandos"`
}

type Disco struct {
	NombreDisco string `json:"NombreDisco"`
}

type ParticionInfo struct {
	Status string `json:"status"`
	Type   string `json:"type"`
	Fit    string `json:"fit"`
	Start  int    `json:"start"`
	Size   int    `json:"size"`
	Name   string `json:"name"`
}

type Response struct {
	Number string `json:"number"`
}

func allowCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*") // Permitir todos los encabezados
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/analizador", analizador).Methods("POST")
	router.HandleFunc("/verficadorDiscos", verificadorDiscos).Methods("GET")
	router.HandleFunc("/getParticiones", getParticiones).Methods("POST")
	handler := allowCORS(router)
	fmt.Println("Se esta escuchando en el puerto 3000")
	log.Fatal(http.ListenAndServe(":3000", handler))
}

func getParticiones(w http.ResponseWriter, r *http.Request) {
	var disco Disco

	// Decodificar el cuerpo de la solicitud en la estructura Disco
	err := json.NewDecoder(r.Body).Decode(&disco)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ahora puedes usar disco.NombreDisco en tu función
	mbr := Comandos.LeerDisco(disco.NombreDisco)
	particiones := Comandos.GetParticiones(*mbr)
	var particionesInfo []ParticionInfo

	for i := 0; i < len(particiones); i++ {
		particion := particiones[i]
		nombre := ""
		if particion.Part_status == "1"[0] {

			for j := 0; j < len(particion.Part_name); j++ {
				if particion.Part_name[j] != 0 {
					nombre += string(particion.Part_name[j])
				}
			}
		}
		particionInfo := ParticionInfo{
			Status: string(particion.Part_status),
			Type:   string(particion.Part_type),
			Fit:    string(particion.Part_fit),
			Start:  int(particion.Part_start),
			Size:   int(particion.Part_size),
			Name:   nombre,
		}
		particionesInfo = append(particionesInfo, particionInfo)
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(particionesInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func verificadorDiscos(w http.ResponseWriter, r *http.Request) {
	path := ""
	CantidadDiscos := 0
	for i := 'A'; i <= 'Z'; i++ {
		path = string(i) + ".dsk"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// File does not exist, create it
			break
		} else {
			CantidadDiscos++
		}
	}
	CD := strconv.Itoa(CantidadDiscos)

	resp := Response{
		Number: CD,
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func analizador(w http.ResponseWriter, r *http.Request) {
	var datos DatosEntrada
	err := json.NewDecoder(r.Body).Decode(&datos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = guardarDatos("./prueba.script", datos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ejecutar el archivo de script
	Exec("./prueba.script")
	fmt.Fprintf(w, "Script ejecutado exitosamente")
}

func guardarDatos(archivo string, datos DatosEntrada) error {
	// Abrir o crear el archivo
	file, err := os.Create(archivo)
	if err != nil {
		return err
	}
	defer file.Close()

	// Escribir los comandos en el archivo
	for _, comando := range datos.Comandos {
		_, err := file.WriteString(strings.TrimSpace(comando) + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func Comando(text string) string {
	var tkn string
	terminar := false
	for i := 0; i < len(text); i++ {
		if terminar {
			if string(text[i]) == " " || string(text[i]) == "-" {
				break
			}
			tkn += string(text[i])
		} else if string(text[i]) != " " && !terminar {
			if string(text[i]) == "#" {
				tkn = text
			} else {
				tkn += string(text[i])
				terminar = true
			}
		}
	}
	return tkn
}

func SepararTokens(texto string) []string {
	var tokens []string
	if texto == "" {
		return tokens
	}
	texto += " "
	var token string
	estado := 0
	for i := 0; i < len(texto); i++ {
		c := string(texto[i])
		if estado == 0 && c == "-" {
			estado = 1
		} else if estado == 0 && c == "#" {
			continue
		} else if estado != 0 {
			if estado == 1 {
				if c == "=" {
					estado = 2
				} else if c == " " {
					continue
				} else if (c == "P" || c == "p") && string(texto[i+1]) == " " && string(texto[i-1]) == "-" {
					estado = 0
					tokens = append(tokens, c)
					token = ""
					continue
				} else if (c == "R" || c == "r") && string(texto[i+1]) == " " && string(texto[i-1]) == "-" {
					estado = 0
					tokens = append(tokens, c)
					token = ""
					continue
				}
			} else if estado == 2 {
				if c == " " {
					continue
				}
				if c == "\"" {
					estado = 3
					continue
				} else {
					estado = 4
				}
			} else if estado == 3 {
				if c == "\"" {
					estado = 4
					continue
				}
			} else if estado == 4 && c == "\"" {
				tokens = []string{}
				continue
			} else if estado == 4 && c == " " {
				estado = 0
				tokens = append(tokens, token)
				token = ""
				continue
			}
			token += c
		}
	}
	return tokens
}

func funciones(token string, tks []string) {
	if token != "" {
		if Comandos.Comparar(token, "EXECUTE") {
			fmt.Println("--------------------------------------- FUNCIÓN EXEC -----------------------------------------")
			FuncionExec(tks)
		} else if Comandos.Comparar(token, "MKDISK") {
			fmt.Println("--------------------------------------- FUNCIÓN MKDISK ---------------------------------------")
			Comandos.ValidarDatosMKDISK(tks)
		} else if Comandos.Comparar(token, "RMDISK") {
			fmt.Println("--------------------------------------- FUNCIÓN RMDISK ---------------------------------------")
			Comandos.RMDISK(tks)
		} else if Comandos.Comparar(token, "FDISK") {
			fmt.Println("--------------------------------------- FUNCIÓN FDISK  ---------------------------------------")
			Comandos.ValidarDatosFDISK(tks)
		} else if Comandos.Comparar(token, "MOUNT") {
			fmt.Println("--------------------------------------- FUNCIÓN MOUNT  ---------------------------------------")
			Comandos.ValidarDatosMOUNT(tks)
		} else if Comandos.Comparar(token, "UNMOUNT") {
			fmt.Println("--------------------------------------- FUNCIÓN UNMOUNT ---------------------------------------")
			Comandos.ValidarDatosUNMOUNT(tks)
		} else if Comandos.Comparar(token, "REP") {
			fmt.Println("--------------------------------------- FUNCIÓN REP ---------------------------------------")
			Reportes.ValidarDatosReporte(tks)
		} else if Comandos.Comparar(token, "MKFS") {
			fmt.Println("--------------------------------------- FUNCIÓN MKFS ---------------------------------------")
			Comandos.ValidarDatosMKFS(tks)
		} else if Comandos.Comparar(token, "LOGIN") {
			fmt.Println("--------------------------------------- FUNCIÓN LOGIN ---------------------------------------")
			if logued {
				Comandos.Error("LOGIN", "Ya hay un usuario en línea.")
				return
			} else {
				logued = Comandos.ValidarDatosLOGIN(tks)
			}

		} else if Comandos.Comparar(token, "LOGOUT") {
			fmt.Println("--------------------------------------- FUNCIÓN LOGOUT ---------------------------------------")
			if !logued {
				Comandos.Error("LOGOUT", "Aún no se ha iniciado sesión.")
				return
			} else {
				logued = Comandos.CerrarSesion()
			}
		} else if Comandos.Comparar(token, "MKGRP") {
			fmt.Println("--------------------------------------- FUNCIÓN MKGRP --------------------------------------- ")
			if !logued {
				Comandos.Error("MKGRP", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosGrupos(tks, "MK")
			}
		} else if Comandos.Comparar(token, "RMGRP") {
			fmt.Println("--------------------------------------- FUNCIÓN RMGRP ---------------------------------------")
			if !logued {
				Comandos.Error("RMGRP", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosGrupos(tks, "RM")
			}
		} else if Comandos.Comparar(token, "MKUSR") {
			fmt.Println("--------------------------------------- FUNCIÓN MKUSER  ---------------------------------------")
			if !logued {
				Comandos.Error("MKUSR", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosUsers(tks, "MK")
			}
		} else if Comandos.Comparar(token, "RMUSR") {
			fmt.Println("--------------------------------------- FUNCIÓN RMUSER ---------------------------------------")
			if !logued {
				Comandos.Error("RMUSR", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosUsers(tks, "RM")
			}
		} else if Comandos.Comparar(token, "MKFILE") {
			fmt.Println("--------------------------------------- FUNCIÓN MKFILE ---------------------------------------")
			if !logued {
				Comandos.Error("MKFILE", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosFile(tks)
			}
		} else if Comandos.Comparar(token, "MKDIR") {
			fmt.Println("--------------------------------------- FUNCIÓN MKDIR ---------------------------------------")
			if !logued {
				Comandos.Error("MKDIR", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosDir(tks)
			}
		} else if Comandos.Comparar(token, "CHOWN") {
			fmt.Println("--------------------------------------- FUNCIÓN CHOWN ---------------------------------------")
			if !logued {
				Comandos.Error("CHOWN", "Aún no se ha iniciado sesión.")
				return
			} else {
				Comandos.ValidarDatosChown(tks)
			}
		} else {
			Comandos.Error("ANALIZADOR", "No se reconoce el comando \""+token+"\"")
		}

	}
}

func FuncionExec(tokens []string) {
	path := ""
	for i := 0; i < len(tokens); i++ {
		datos := strings.Split(tokens[i], "=")
		if Comandos.Comparar(datos[0], "path") {
			path = datos[1]
		}
	}
	if path == "" {
		Comandos.Error("EXEC", "Se requiere el parámetro \"path\" para este comando")
		return
	}
	Exec(path)
}

func Exec(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error al abrir el archivo: %s", err)
	}
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		texto := fileScanner.Text()
		texto = strings.TrimSpace(texto)
		tk := Comando(texto)
		if texto != "" {
			if Comandos.Comparar(tk, "pause") {
				fmt.Println("--------------------------------------- FUNCIÓN PAUSE ---------------------------------------")
				var pause string
				Comandos.Mensaje("PAUSE", "Presione \"enter\" para continuar...")
				fmt.Scanln(&pause)
				continue
			} else if string(texto[0]) == "#" {
				fmt.Println("--------------------------------------- COMENTARIO ------------------------------------------")
				Comandos.Mensaje("COMENTARIO", texto)
				continue
			}
			texto = strings.TrimLeft(texto, tk)
			tokens := SepararTokens(texto)
			funciones(tk, tokens)
		}
	}
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error al leer el archivo: %s", err)
	}
	file.Close()
}
