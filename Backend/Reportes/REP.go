package Reportes

import (
	"Proyecto/Comandos"
	"Proyecto/Structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"

	"github.com/awalterschulze/gographviz"
)

type Apuntadores struct {
	Inodos    []string
	Bloques   []string
	Direccion []string
}

type CarpetaFront struct {
	NombreCarpeta string
}

type ArchivoFront struct {
	NumArchivo int
	Contenido  string
}

type ContenidoFront struct {
	Carpetas []CarpetaFront
	Archivos []ArchivoFront
}

var Ap = Apuntadores{}

func ValidarDatosReporte(context []string) {
	name := ""
	path := ""
	id := ""
	ruta := ""
	for i := 0; i < len(context); i++ {
		current := context[i]
		comando := strings.Split(current, "=")
		if Comandos.Comparar(comando[0], "name") {
			name = comando[1]
		} else if Comandos.Comparar(comando[0], "path") {
			path = strings.ReplaceAll(comando[1], "\"", "")
		} else if Comandos.Comparar(comando[0], "id") {
			id = comando[1]
		} else if Comandos.Comparar(comando[0], "ruta") {
			ruta = comando[1]
		}
	}
	if path == "" || name == "" || id == "" {
		Comandos.Error("REP", "Requiere parámetros obligatorios")
		return
	}

	if Comandos.Comparar(name, "mbr") {
		MBR_R(path, id)
		return
	} else if Comandos.Comparar(name, "disk") {
		DISK_R(path, id, ruta)
		return
	} else if Comandos.Comparar(name, "tree") {
		ReporteTree(path, id)
		return
	} else if Comandos.Comparar(name, "bm_inode") {
		BitMap_inodo(path, id)
	} else if Comandos.Comparar(name, "bm_block") {
		BitMap_block(path, id)
	} else if Comandos.Comparar(name, "inode") {
		Report_Inode(path, id)
	} else if Comandos.Comparar(name, "block") {
		Report_Block(path, id)
	} else if Comandos.Comparar(name, "sb") {
		SB_Reporte(path, id)
	} else if Comandos.Comparar(name, "Journaling") {
		Journal_Reporte(path, id)
	} else {
		Comandos.Error("REP", "Nombre de reporte no válido")
		return
	}

}

func Journal_Reporte(destino string, id string) {
	Ap.Bloques = nil
	Ap.Inodos = nil
	Ap.Direccion = nil
	fmt.Println("Generando reporte de Jornal")
	path := id[0:1]
	path = path + ".dsk"

	partcion := Comandos.GetMount("REP", id, &path)
	spr := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Comandos.Error("REP", "No se ha encontrado el disco.")
		return
	}
	tamJournal := int64(unsafe.Sizeof(Structs.Journaling{}))

	file.Seek(partcion.Part_start, 0)
	data := lecturaB(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &spr)
	if err_ != nil {
		Comandos.Error("REP", "Error al leer el archivo")
		return
	}
	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}
	InicioJournal := partcion.Part_start + int64(unsafe.Sizeof(Structs.SuperBloque{}))
	Codigo_HTML := "<<TABLE>\n"
	Codigo_HTML += fmt.Sprintf(`
			<TR style="background-color: gray; color: white;">
				<TD BGCOLOR="#006400">
					<FONT COLOR="white">Operacion</FONT>
				</TD>
				<TD BGCOLOR="#006400">
					<FONT COLOR="white">Path</FONT>
				</TD>
				<TD BGCOLOR="#006400">
					<FONT COLOR="white">Contenido</FONT>
				</TD>
				<TD BGCOLOR="#006400">
					<FONT COLOR="white">Fecha</FONT>
				</TD>
			</TR>`)
	for {
		tempJournal := Structs.NewJournal()
		file.Seek(InicioJournal, 0)
		data := lecturaB(file, int(unsafe.Sizeof(Structs.Journaling{})))
		buffer := bytes.NewBuffer(data)
		err_ := binary.Read(buffer, binary.BigEndian, &tempJournal)
		if err_ != nil {
			Comandos.Error("REP", "Error al leer el archivo")
			return
		}

		// Primera condición
		if tempJournal.Journaling_start != 0 {

			Operacion := string(bytes.Trim(tempJournal.Operacion[:], "\x00"))
			Fecha := string(bytes.Trim(tempJournal.Fecha[:], "\x00"))
			Path := string(bytes.Trim(tempJournal.Path[:], "\x00"))
			Contenido := string(bytes.Trim(tempJournal.Contenido[:], "\x00"))
			Codigo_HTML += fmt.Sprintf(`
			<TR>
				<TD>%s</TD>
				<TD>%s</TD>
				<TD>%s</TD>
				<TD>%s</TD>
			</TR>
			`, Operacion, Path, Contenido, Fecha)
			InicioJournal += tamJournal

		} else {
			break
		}
	}

	Codigo_HTML += fmt.Sprintf(`</TABLE>>`)
	graph.AddNode("G", "a", map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
	Rdot := destino + ".dot"
	// Guarda el código DOT en un archivo
	err = ioutil.WriteFile(Rdot, []byte(graph.String()), 0644)
	if err != nil {
		fmt.Println(err)
	}

	R := destino + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

}

func SB_Reporte(destino string, id string) {
	Ap.Bloques = nil
	Ap.Inodos = nil
	Ap.Direccion = nil
	fmt.Println("Generando reporte de superbloque")
	path := id[0:1]
	path = path + ".dsk"

	partcion := Comandos.GetMount("REP", id, &path)
	spr := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Comandos.Error("REP", "No se ha encontrado el disco.")
		return
	}

	file.Seek(partcion.Part_start, 0)
	data := lecturaB(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &spr)
	if err_ != nil {
		Comandos.Error("REP", "Error al leer el archivo")
		return
	}
	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}
	Codigo_HTML := fmt.Sprintf(`
<<TABLE>
    <TR style="background-color: #006400; color: white;">
    <TD BGCOLOR="#006400">
        <FONT COLOR="white">Reporte SuperBloque</FONT>
    </TD>
    <TD BGCOLOR="#006400">
        <FONT COLOR="white">Datos</FONT>
    </TD>
	</TR>
    <TR>
        <TD>S_filesystem_type</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_inodes_count</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_blocks_count</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_free_blocks_count</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_free_inodes_count</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_mtime</TD><TD>%s</TD>
    </TR>
    <TR>
        <TD>S_umtime</TD><TD>%s</TD>
    </TR>
    <TR>
        <TD>S_mnt_count</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_magic</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_inode_size</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_block_size</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_firts_ino</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_first_blo</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_bm_inode_start</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_bm_block_start</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_inode_start</TD><TD>%d</TD>
    </TR>
    <TR>
        <TD>S_block_start</TD><TD>%d</TD>
    </TR>
`, spr.S_filesystem_type, spr.S_inodes_count, spr.S_blocks_count, spr.S_free_blocks_count, spr.S_free_inodes_count, string(spr.S_mtime[:]), string(spr.S_umtime[:]), spr.S_mnt_count, spr.S_magic, spr.S_inode_size, spr.S_block_size, spr.S_firts_ino, spr.S_first_blo, spr.S_bm_inode_start, spr.S_bm_block_start, spr.S_inode_start, spr.S_block_start)
	Codigo_HTML += fmt.Sprintf(`</TABLE>>`)
	graph.AddNode("G", "a", map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
	Rdot := destino + ".dot"
	// Guarda el código DOT en un archivo
	err = ioutil.WriteFile(Rdot, []byte(graph.String()), 0644)
	if err != nil {
		fmt.Println(err)
	}

	R := destino + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

}

func MBR_R(path string, id string) {

	primeraLetra := id[0:1]
	primeraLetra = primeraLetra + ".dsk"

	mbr := Comandos.LeerDisco(string(primeraLetra))

	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}

	// Añade un nodo como tabla

	mbrTamano := strconv.FormatInt(mbr.Mbr_tamano, 10)
	mbrFechaCreacion := mbr.Mbr_fecha_creacion
	mbrDiskSignature := strconv.FormatInt(mbr.Mbr_dsk_signature, 10)

	Codigo_HTML := fmt.Sprintf(`<<TABLE>
    <TR style="background-color: #4B0082; color: white;">
        <TD BGCOLOR="#4B0082">
            <FONT COLOR="white">Reporte MBR</FONT>
        </TD>
        <TD BGCOLOR="#4B0082">
            <FONT COLOR="white">Datos</FONT>
        </TD>
    </TR>
    <TR>
        <TD>Mbr_tamano</TD><TD>%s</TD>
    </TR>
    <TR>
        <TD>Mbr_fecha_creacion</TD><TD>%s</TD>
    </TR>
    <TR>
        <TD>Mbr_disk_signature</TD><TD>%s</TD>
    </TR>
	`, mbrTamano, mbrFechaCreacion, mbrDiskSignature)
	particiones := Comandos.GetParticiones(*mbr)
	for i := 0; i < len(particiones); i++ {
		particion := particiones[i]
		Codigo_HTML += fmt.Sprintf(`
		<TR style="background-color: #4B0082; color: white;">
			<TD BGCOLOR="#4B0082">
				<FONT COLOR="white">Particion</FONT>
			</TD>
			<TD BGCOLOR="#4B0082">
            	
        	</TD>
		</TR>
		`)
		if particion.Part_type != "E"[0] {

			TipoParticion := string(particion.Part_type)
			PartStatus := string(particion.Part_status) // Convert byte to string
			PartStart := strconv.FormatInt(particion.Part_start, 10)
			PartSize := strconv.FormatInt(particion.Part_size, 10)
			PartFit := string(particion.Part_fit)

			PartName := ""
			for _, b := range particion.Part_name {
				if b != 0 {
					PartName += string(b)
				} else {
					break
				}
			}
			Codigo_HTML += fmt.Sprintf(`

			<TR>
				<TD>Part_status</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_type</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_fit</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_start</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_size</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_name</TD><TD>%s</TD>
			</TR>
		`, PartStatus, TipoParticion, PartFit, PartStart, PartSize, PartName)
		} else if particion.Part_type == "E"[0] || particion.Part_type == "e"[0] {
			TipoParticion := string(particion.Part_type)
			PartStatus := string(particion.Part_status) // Convert byte to string
			PartStart := strconv.FormatInt(particion.Part_start, 10)
			PartSize := strconv.FormatInt(particion.Part_size, 10)
			PartFit := string(particion.Part_fit)

			PartName := ""
			for _, b := range particion.Part_name {
				if b != 0 {
					PartName += string(b)
				} else {
					break
				}
			}
			Codigo_HTML += fmt.Sprintf(`

			<TR>
				<TD>Part_status</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_type</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_fit</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_start</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_size</TD><TD>%s</TD>
			</TR>
			<TR>
				<TD>Part_name</TD><TD>%s</TD>
			</TR>
			`, PartStatus, TipoParticion, PartFit, PartStart, PartSize, PartName)
			// Buscar particiones lógicas
			ebrs := Comandos.GetLogicas(particion, primeraLetra)
			for i := 0; i < len(ebrs); i++ {
				Codigo_HTML += fmt.Sprintf(`
				<TR style="background-color: #FA8072; color: white;">
					<TD BGCOLOR="#FA8072">
						<FONT COLOR="white">Partición logica</FONT>
					</TD>
					<TD BGCOLOR="#FA8072">
						
					</TD>
				</TR>
				`)
				ebr := ebrs[i]
				TipoParticion_Logica := string("L")
				PartStatus_Logica := string(ebr.Part_status) // Convert byte to string
				PartStart_Logica := strconv.FormatInt(ebr.Part_start, 10)
				PartSize_Logica := strconv.FormatInt(ebr.Part_size, 10)
				PartFit_Logica := string(particion.Part_fit)

				PartName_Logica := ""

				for j := 0; j < len(ebr.Part_name); j++ {
					if ebr.Part_name[j] != 0 {
						PartName_Logica += string(ebr.Part_name[j])
					}
				}

				Codigo_HTML += fmt.Sprintf(`

				<TR>
					<TD>Part_status</TD><TD>%s</TD>
				</TR>
				<TR>
					<TD>Part_type</TD><TD>%s</TD>
				</TR>
				<TR>
					<TD>Part_fit</TD><TD>%s</TD>
				</TR>
				<TR>
					<TD>Part_start</TD><TD>%s</TD>
				</TR>
				<TR>
					<TD>Part_size</TD><TD>%s</TD>
				</TR>
				<TR>
					<TD>Part_name</TD><TD>%s</TD>
				</TR>
			`, PartStatus_Logica, TipoParticion_Logica, PartFit_Logica, PartStart_Logica, PartSize_Logica, PartName_Logica)
			}
		}
	}
	Codigo_HTML += fmt.Sprintf(`</TABLE>>`)
	graph.AddNode("G", "a", map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
	Rdot := path + ".dot"
	// Guarda el código DOT en un archivo
	err := ioutil.WriteFile(Rdot, []byte(graph.String()), 0644)
	if err != nil {
		fmt.Println(err)
	}

	R := path + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func DISK_R(path string, id string, ruta string) {
	fmt.Println("Generando reporte DISK")

	primeraLetra := id[0:1]
	primeraLetra = primeraLetra + ".dsk"
	mbr := Comandos.LeerDisco(string(primeraLetra))
	particiones := Comandos.GetParticiones(*mbr)
	extended := Structs.NewParticion()
	tam_disk := int(mbr.Mbr_tamano)
	tam_libre := tam_disk

	Codigo_HTML := fmt.Sprintf(`<TABLE BORDER="0" CELLBORDER="3" CELLSPACING="0">
	               <TR>
	                   <TD BGCOLOR="lightblue"><FONT POINT-SIZE="20">MBR</FONT></TD>
	               `)
	puntero := int(float64(tam_disk) * 0.1)
	for i := 0; i < len(particiones); i++ {
		particion := particiones[i]
		if particion.Part_status == "1"[0] {
			if particion.Part_type == "E"[0] || particion.Part_type == "e"[0] {
				extended = particion
				tamano := int(particion.Part_size)
				puntero += int(particion.Part_size)
				tam_libre -= int(particion.Part_size)
				tamano = int((float64(tamano) / float64(tam_disk)) * 100)
				Codigo_HTML += fmt.Sprintf(`
				<TD>
					<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
						<TR>
					<TD BGCOLOR="green"><FONT COLOR="black">Particion extendida %s <BR/></FONT>
					<FONT POINT-SIZE="10">Tamaño: %s  %% </FONT>
					</TD>
				`, strconv.Itoa(i+1), strconv.Itoa(tamano))

				ebrs := Comandos.GetLogicas(extended, primeraLetra)
				for i := 0; i < len(ebrs); i++ {
					ebr := ebrs[i]
					if ebr.Part_status == '1' {

						tamano := int(ebr.Part_size)
						tam_libre -= int(ebr.Part_size)
						tamano = int((float64(tamano) / float64(tam_disk)) * 100)
						Codigo_HTML += fmt.Sprintf(`<TD BGCOLOR="#FA8072">Particion logica %s <BR/>
							<FONT POINT-SIZE="7">Tamaño: %s  %% </FONT>
							</TD>
						`, strconv.Itoa(i+1), strconv.Itoa(tamano))

					}
				}
				Codigo_HTML += fmt.Sprintf(`</TR>
					</TABLE>
				</TD>`)
			} else if particion.Part_type == "P"[0] || particion.Part_type == "p"[0] {
				tamano := int(particion.Part_size)
				tam_libre -= int(particion.Part_size)
				tamano = int((float64(tamano) / float64(tam_disk)) * 100)

				Codigo_HTML += fmt.Sprintf(`
				<TD BGCOLOR="#4B0082">
					<FONT COLOR="white">Particion primaria %s </FONT> <BR/><FONT POINT-SIZE="10">Tamaño: %s  %% </FONT></TD>
				`, strconv.Itoa(i+1), strconv.Itoa(tamano))

			}
		} else {
			// convierte i a string
			if i == 0 {
				fmt.Println("Particion 1 libre")
				if particiones[i+1].Part_start == -1 {
					tamano := tam_disk - puntero
					tamano = int((float64(tamano) / float64(tam_disk)) * 100)
					Codigo_HTML += fmt.Sprintf(`
					<TD BGCOLOR="White">
						<FONT COLOR="black">Espacio libre </FONT> <BR/><FONT POINT-SIZE="10">Tamaño: %s  %% </FONT></TD>
					`, strconv.Itoa(tamano))
					break
				}
			} else if i == 1 {
				fmt.Println("Particion 2 libre")
				if particiones[i+1].Part_start == -1 {
					tamano := tam_disk - puntero
					tamano = int((float64(tamano) / float64(tam_disk)) * 100)
					Codigo_HTML += fmt.Sprintf(`
					<TD BGCOLOR="White">
						<FONT COLOR="black">Espacio libre </FONT> <BR/><FONT POINT-SIZE="10">Tamaño: %s  %% </FONT></TD>
					`, strconv.Itoa(tamano))
					break
				}

			} else if i == 2 {
				fmt.Println("Particion 3 libre")

				if particiones[i+1].Part_start == -1 {
					tamano := tam_disk - puntero
					tamano = int((float64(tamano) / float64(tam_disk)) * 100)
					Codigo_HTML += fmt.Sprintf(`
					<TD BGCOLOR="White">
						<FONT COLOR="black">Espacio libre </FONT> <BR/><FONT POINT-SIZE="10">Tamaño: %s  %% </FONT></TD>
					`, strconv.Itoa(tamano))
					break
				} else {
					tamano := int(particiones[i+1].Part_start) - puntero
					tamano = int((float64(tamano) / float64(tam_disk)) * 100)
					Codigo_HTML += fmt.Sprintf(`
					<TD BGCOLOR="#4B0082">
						<FONT COLOR="white">Espacio libre </FONT> <BR/><FONT POINT-SIZE="10">Tamaño: %s  %% </FONT></TD>
					`, strconv.Itoa(tamano))
				}

			} else if i == 3 {
				fmt.Println("Particion 4 libre")
				tamano := tam_disk - puntero
				tamano = int((float64(tamano) / float64(tam_disk)) * 100)
				Codigo_HTML += fmt.Sprintf(`
					<TD BGCOLOR="White">
						<FONT COLOR="black">Espacio libre </FONT> <BR/><FONT POINT-SIZE="10">Tamaño: %s  %% </FONT></TD>
					`, strconv.Itoa(tamano))
				break

			}

		}
	}

	Codigo_HTML += fmt.Sprintf(`
			</TR>
		</TABLE>`)

	dotContent := fmt.Sprintf(`digraph G {
	       rankdir=LR;
	       node [shape=none];
	       DiscoDuro [label=<%s>];
	   }`, Codigo_HTML)
	Rdot := path + ".dot"
	// Guarda el código DOT en un archivo
	err := ioutil.WriteFile(Rdot, []byte(dotContent), 0644)
	if err != nil {
		fmt.Println(err)
	}
	R := path + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

}

func ReporteTree(destino string, id string) {

	Ap.Inodos = []string{}
	Ap.Bloques = []string{}
	Ap.Direccion = []string{}
	TamBloqueCarpeta := int(unsafe.Sizeof(Structs.BloquesCarpetas{}))
	TamBloqueArchivo := int(unsafe.Sizeof(Structs.BloquesArchivos{}))
	TamInodo := int(unsafe.Sizeof(Structs.Inodos{}))

	fmt.Println("Generando reporte de tree")
	path := id[0:1]
	path = path + ".dsk"

	partcion := Comandos.GetMount("REP", id, &path)
	SB := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Comandos.Error("REP", "No se ha encontrado el disco.")
		return
	}

	file.Seek(partcion.Part_start, 0)
	data := lecturaB(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &SB)
	if err_ != nil {
		Comandos.Error("REP", "Error al leer el archivo")
		return
	}

	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	graph.Attrs.Add("rankdir", "LR")
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}
	MitadBA := (partcion.Part_size - SB.S_block_start) / 2
	MitadBA = MitadBA + SB.S_block_start

	Codigo_HTML := ""
	inode := Structs.NewInodos()
	PunteroInodos := SB.S_inode_start
	PunteroBloquesCarpetas := SB.S_block_start
	PunteroBloquesArchivos := MitadBA
	//CantidadBloques := 0

	CantidadBloquesCarpetas := 0
	CantidadBloquesArchivos := 0

	for {
		bc := Structs.NewBloquesCarpetas()
		file.Seek(PunteroBloquesCarpetas, 0)
		data = lecturaB(file, TamBloqueCarpeta)
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &bc)
		if err_ != nil {
			Comandos.Error("MkDir", "Error al leer el archivo")
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

	for {
		var fb Structs.BloquesArchivos
		file.Seek(PunteroBloquesArchivos, 0)
		data = lecturaB(file, TamBloqueArchivo)
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &fb)
		if err_ != nil {
			Comandos.Error("REP", "Error al leer el archivo")
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

	fmt.Println("Cantidad de bloques carpetas: ", CantidadBloquesCarpetas)
	fmt.Println("Cantidad de bloques archivos: ", CantidadBloquesArchivos)

	BCN := 0
	BA := 0

	for i := 0; i < int(SB.S_inodes_count); i++ {
		Carpeta := true

		file.Seek(PunteroInodos, 0)
		data = lecturaB(file, TamInodo)
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &inode)
		if err_ != nil {
			Comandos.Error("REP", "Error al leer el archivo")
			return
		}

		if inode.I_size != -1 {
			labelInode := fmt.Sprintf("Inodo%d", i)
			Ap.Inodos = append(Ap.Inodos, labelInode)
			Codigo_HTML = GenerateInodo(inode, i)
			graph.AddNode("G", labelInode, map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
			//Carpeta 0 Archivo 1
			if inode.I_type == 1 {
				Carpeta = false
			} else if inode.I_type > 1 {
				fmt.Println("No valido")
			}

			for j := 0; j < len(inode.I_block); j++ {
				if inode.I_block[j] != -1 {
					Apunta := int(inode.I_block[j])
					if Carpeta == true {

						bc := Structs.NewBloquesCarpetas()

						if Apunta == 0 {
							PunteroBloquesCarpetas = SB.S_block_start + (int64(Apunta) * int64(TamBloqueCarpeta))
						} else {
							resut := int64(Apunta - 2)
							fmt.Println("Resultado: ", resut)
							PunteroBloquesCarpetas = SB.S_block_start + resut*int64(TamBloqueCarpeta)
						}

						file.Seek(PunteroBloquesCarpetas, 0)
						data = lecturaB(file, TamBloqueCarpeta)
						buffer = bytes.NewBuffer(data)
						err_ = binary.Read(buffer, binary.BigEndian, &bc)
						if err_ != nil {

							Comandos.Error("REP", "Error al leer el archivo")
							return
						}
						labelBloque := fmt.Sprintf("Bloque%d", Apunta)
						Ap.Bloques = append(Ap.Bloques, labelBloque)
						Codigo_HTML = GenerateBloqueCarpetas(bc, i, labelBloque, Apunta)
						graph.AddNode("G", labelBloque, map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
						BCN++

					} else {
						if Apunta < 3 {
							PunteroBloquesArchivos = MitadBA + (int64(Apunta-1) * int64(TamBloqueArchivo))
						} else {
							PunteroBloquesArchivos = MitadBA + (int64(Apunta-CantidadBloquesCarpetas) * int64(TamBloqueArchivo))
						}

						var fb Structs.BloquesArchivos
						file.Seek(PunteroBloquesArchivos, 0)
						data = lecturaB(file, TamBloqueArchivo)
						buffer = bytes.NewBuffer(data)
						err_ = binary.Read(buffer, binary.BigEndian, &fb)
						if err_ != nil {
							Comandos.Error("REP", "Error al leer el archivo")
							return
						}

						txt := ""
						for i := 0; i < len(fb.B_content); i++ {
							if fb.B_content[i] != 0 {
								txt += string(fb.B_content[i])
							}
							if len(txt) == 64 {
								break
							}
						}
						labelBloque := fmt.Sprintf("Bloque%d", Apunta)
						Ap.Bloques = append(Ap.Bloques, labelBloque)
						Codigo_HTML = GenerateBloqueArchivo(txt, i, Apunta)
						graph.AddNode("G", labelBloque, map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
						BA++
					}

				}
			}
		} else {
			break
		}
		PunteroInodos += int64(TamInodo)
	}
	graphParts := strings.Split(graph.String(), "}\n")
	Direccionamiento := ""
	for i := 0; i < len(Ap.Direccion); i++ {
		Direccionamiento += "\t" + Ap.Direccion[i] + "\n"
	}

	graphRaw := graphParts[0] + Direccionamiento + "\n}"

	Rdot := destino + ".dot"
	// Guarda el código DOT en un archivo
	errdot := ioutil.WriteFile(Rdot, []byte(graphRaw), 0644)
	if errdot != nil {
		fmt.Println(errdot)
	}

	R := destino + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	errdot = cmd.Run()
	if errdot != nil {
		fmt.Println(errdot)
	}

}

func BitMap_inodo(destino string, id string) {
	fmt.Println("Generando reporte de Bitmap inodos")
	path := id[0:1]
	path = path + ".dsk"
	partcion := Comandos.GetMount("REP", id, &path)
	SB := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Comandos.Error("REP", "No se ha encontrado el disco.")
		return
	}
	file.Seek(partcion.Part_start, 0)
	data := lecturaB(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &SB)
	if err_ != nil {
		Comandos.Error("REP", "Error al leer el archivo")
		return
	}
	ceros := 0
	unos := 0

	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}

	Codigo_HTML := fmt.Sprintf(`<<TABLE>
	`)

	control := 40
	Codigo_HTML += fmt.Sprintf(`<TR style="background-color: #4B0082; color: white;">`)
	for i := 0; i < int(SB.S_inodes_count); i++ {
		if control == 0 {
			Codigo_HTML += fmt.Sprintf(`</TR>`)
			Codigo_HTML += fmt.Sprintf(`<TR style="background-color: #4B0082; color: white;">`)
			control = 40
		}

		file.Seek(SB.S_bm_inode_start+int64(i), 0)
		data = lecturaB(file, 1)
		buffer = bytes.NewBuffer(data)
		var bm byte
		err_ = binary.Read(buffer, binary.BigEndian, &bm)
		if bm == 49 {
			Codigo_HTML += fmt.Sprintf(`<TD BGCOLOR="black" STYLE="color: lightblue;"><FONT color="lightblue">1</FONT></TD>`)
			unos += 1
		} else if bm == 0 || bm == 48 {
			Codigo_HTML += fmt.Sprintf(`<TD BGCOLOR="black" STYLE="color: lightblue;"><FONT color="lightblue">0</FONT></TD>`)
			ceros += 1
		}
		control -= 1
	}
	Codigo_HTML += fmt.Sprintf(`
	</TR>`)
	Codigo_HTML += fmt.Sprintf(`
	</TABLE>>`)

	graph.AddNode("G", "BitMapInodo", map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
	Rdot := destino + ".dot"
	// Guarda el código DOT en un archivo
	err = ioutil.WriteFile(Rdot, []byte(graph.String()), 0644)
	if err != nil {
		fmt.Println(err)
	}

	R := destino + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

}

func BitMap_block(destino string, id string) {
	fmt.Println("Generando reporte de bitmap bloques")
	path := id[0:1]
	path = path + ".dsk"
	partcion := Comandos.GetMount("REP", id, &path)
	SB := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Comandos.Error("REP", "No se ha encontrado el disco.")
		return
	}
	file.Seek(partcion.Part_start, 0)
	data := lecturaB(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &SB)
	if err_ != nil {
		Comandos.Error("REP", "Error al leer el archivo")
		return
	}
	ceros := 0
	unos := 0

	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}

	Codigo_HTML := fmt.Sprintf(`<<TABLE>
	`)

	control := 40
	Codigo_HTML += fmt.Sprintf(`<TR style="background-color: #4B0082; color: white;">`)
	for i := 0; i < int(SB.S_blocks_count); i++ {
		if control == 0 {
			Codigo_HTML += fmt.Sprintf(`</TR>`)
			Codigo_HTML += fmt.Sprintf(`<TR style="background-color: #4B0082; color: white;">`)
			control = 40
		}

		file.Seek(SB.S_bm_block_start+int64(i), 0)
		data = lecturaB(file, 1)
		buffer = bytes.NewBuffer(data)
		var bm byte
		err_ = binary.Read(buffer, binary.BigEndian, &bm)
		if bm == 49 {
			Codigo_HTML += fmt.Sprintf(`<TD BGCOLOR="black" STYLE="color: lightblue;"><FONT color="lightblue">1</FONT></TD>`)
			unos += 1
		} else if bm == 0 || bm == 48 {
			Codigo_HTML += fmt.Sprintf(`<TD BGCOLOR="black" STYLE="color: lightblue;"><FONT color="lightblue">0</FONT></TD>`)
			ceros += 1
		}
		control -= 1
	}
	Codigo_HTML += fmt.Sprintf(`
	</TR>`)
	Codigo_HTML += fmt.Sprintf(`
	</TABLE>>`)

	graph.AddNode("G", "BitMapBlock", map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
	Rdot := destino + ".dot"
	// Guarda el código DOT en un archivo
	err = ioutil.WriteFile(Rdot, []byte(graph.String()), 0644)
	if err != nil {
		fmt.Println(err)
	}

	R := destino + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

}

func Report_Inode(destino string, id string) {
	fmt.Println("Generando reporte de inodos")
	path := id[0:1]
	path = path + ".dsk"
	Ap.Inodos = []string{}
	Ap.Bloques = []string{}
	Ap.Direccion = []string{}
	partcion := Comandos.GetMount("REP", id, &path)
	SB := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Comandos.Error("REP", "No se ha encontrado el disco.")
		return
	}

	file.Seek(partcion.Part_start, 0)
	data := lecturaB(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &SB)
	if err_ != nil {
		Comandos.Error("REP", "Error al leer el archivo")
		return
	}

	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()

	graph.Attrs.Add("rankdir", "LR")
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}

	Codigo_HTML := ""
	inode := Structs.NewInodos()

	for i := 0; i < int(SB.S_inodes_count); i++ {

		file.Seek(SB.S_inode_start+(int64(unsafe.Sizeof(Structs.Inodos{}))*int64(i)), 0)
		data = lecturaB(file, int(unsafe.Sizeof(Structs.Inodos{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &inode)
		if err_ != nil {
			Comandos.Error("REP", "Error al leer el archivo")
			return
		}
		if inode.I_size != -1 {
			labelInode := fmt.Sprintf("Inodo%d", i)
			Ap.Inodos = append(Ap.Inodos, labelInode)
			Codigo_HTML = GenerateInodo(inode, i)
			graph.AddNode("G", labelInode, map[string]string{"label": Codigo_HTML, "shape": "plaintext"})

		} else {
			break
		}
	}

	Ap.Direccion = []string{}
	for i := 0; i < len(Ap.Inodos); i++ {
		if i+1 < len(Ap.Inodos) {
			Ap.Direccion = append(Ap.Direccion, fmt.Sprintf("%s:i%d -> %s:i%d", Ap.Inodos[i], i, Ap.Inodos[i+1], i+1))
		} else {
			break
		}
	}

	graphParts := strings.Split(graph.String(), "}\n")
	Direccionamiento := ""
	for i := 0; i < len(Ap.Direccion); i++ {
		Direccionamiento += "\t" + Ap.Direccion[i] + "\n"
	}

	graphRaw := graphParts[0] + Direccionamiento + "\n}"
	Rdot := destino + ".dot"
	// Guarda el código DOT en un archivo
	errdot := ioutil.WriteFile(Rdot, []byte(graphRaw), 0644)
	if errdot != nil {
		fmt.Println(errdot)
	}

	R := destino + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	errdot = cmd.Run()
	if errdot != nil {
		fmt.Println(errdot)
	}
}

func Report_Block(destino string, id string) {

	Ap.Bloques = []string{}
	Ap.Inodos = []string{}
	Ap.Direccion = []string{}

	fmt.Println("Generando reporte de bloques")
	path := id[0:1]
	path = path + ".dsk"

	partcion := Comandos.GetMount("REP", id, &path)
	SB := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Comandos.Error("REP", "No se ha encontrado el disco.")
		return
	}

	file.Seek(partcion.Part_start, 0)
	data := lecturaB(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &SB)
	if err_ != nil {
		Comandos.Error("REP", "Error al leer el archivo")
		return
	}

	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	graph := gographviz.NewGraph()
	graph.Attrs.Add("rankdir", "LR")
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		panic(err)
	}

	Codigo_HTML := ""
	inode := Structs.NewInodos()

	BloquesUsados := 0

	for i := 0; i < int(SB.S_inodes_count); i++ {

		file.Seek(SB.S_inode_start+(int64(unsafe.Sizeof(Structs.Inodos{}))*int64(i)), 0)
		data = lecturaB(file, int(unsafe.Sizeof(Structs.Inodos{})))
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &inode)
		if err_ != nil {
			Comandos.Error("REP", "Error al leer el archivo")
			return
		}
		if inode.I_size != -1 {
			labelInode := fmt.Sprintf("Inodo%d", i)

			for j := 0; j < len(inode.I_block); j++ {
				if j < 14 {
					if inode.I_block[j] != -1 {

						if labelInode == "Inodo1" {
							blocArch := int(inode.I_block[j]) - 1
							var fb Structs.BloquesArchivos
							file.Seek(SB.S_block_start+int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))+int64(unsafe.Sizeof(Structs.BloquesArchivos{}))*int64(blocArch), 0)

							data = lecturaB(file, int(unsafe.Sizeof(Structs.BloquesArchivos{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &fb)

							if err_ != nil {
								Comandos.Error("REP", "Error al leer el archivo")
								return
							}
							txt := ""
							for i := 0; i < len(fb.B_content); i++ {
								if fb.B_content[i] != 0 {
									txt += string(fb.B_content[i])
								}
								if len(txt) == 64 {
									break
								}
							}

							labelBloque := fmt.Sprintf("Bloque%d", BloquesUsados)
							Ap.Bloques = append(Ap.Bloques, labelBloque)
							Codigo_HTML = GenerateBloqueArchivo(txt, int(inode.I_gid), BloquesUsados)
							graph.AddNode("G", labelBloque, map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
							BloquesUsados += 1

						} else {

							bc := Structs.NewBloquesCarpetas()
							file.Seek(SB.S_block_start+(int64(unsafe.Sizeof(Structs.BloquesCarpetas{}))*int64(inode.I_block[j])), 0)
							data = lecturaB(file, int(unsafe.Sizeof(Structs.BloquesCarpetas{})))
							buffer = bytes.NewBuffer(data)
							err_ = binary.Read(buffer, binary.BigEndian, &bc)
							if err_ != nil {
								Comandos.Error("REP", "Error al leer el archivo")
								return
							}
							labelBloque := fmt.Sprintf("Bloque%d", BloquesUsados)
							Ap.Bloques = append(Ap.Bloques, labelBloque)
							Codigo_HTML = GenerateBloqueCarpetas(bc, int(inode.I_gid), labelBloque, BloquesUsados)
							graph.AddNode("G", labelBloque, map[string]string{"label": Codigo_HTML, "shape": "plaintext"})
							BloquesUsados += 1
						}

					}
				}
			}
		} else {
			break
		}
	}

	Ap.Direccion = []string{}
	for i := 0; i < len(Ap.Bloques); i++ {
		if i+1 < len(Ap.Bloques) {
			Ap.Direccion = append(Ap.Direccion, fmt.Sprintf("%s:i%d -> %s:i%d", Ap.Bloques[i], i, Ap.Bloques[i+1], i+1))
		} else {
			break
		}
	}

	graphParts := strings.Split(graph.String(), "}\n")
	Direccionamiento := ""
	for i := 0; i < len(Ap.Direccion); i++ {
		Direccionamiento += "\t" + Ap.Direccion[i] + "\n"
	}

	graphRaw := graphParts[0] + Direccionamiento + "\n}"
	Rdot := destino + ".dot"
	// Guarda el código DOT en un archivo
	errdot := ioutil.WriteFile(Rdot, []byte(graphRaw), 0644)
	if errdot != nil {
		fmt.Println(errdot)
	}

	R := destino + ".png"
	// Genera el archivo PNG usando la herramienta dot
	cmd := exec.Command("dot", "-Tpng", Rdot, "-o", R)
	errdot = cmd.Run()
	if errdot != nil {
		fmt.Println(errdot)
	}

}

func GenerateInodo(inodo Structs.Inodos, numInodo int) string {
	retorno := fmt.Sprintf(`
        <<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" BGCOLOR="lightblue">
            <TR><TD COLSPAN="2" PORT="i%d"><B>Inode %d</B></TD></TR>
            <TR><TD>I_uid</TD><TD>%d</TD></TR>
            <TR><TD>I_gid</TD><TD>%d</TD></TR>
            <TR><TD>I_size</TD><TD>%d</TD></TR>
            <TR><TD>I_type</TD><TD>%d</TD></TR>
            <TR><TD>I_perm</TD><TD>%d</TD></TR>
        `, numInodo, numInodo, inodo.I_uid, inodo.I_gid, inodo.I_size, inodo.I_type, inodo.I_perm)

	for i := 0; i < 16; i++ {

		if inodo.I_block[i] != -1 {
			port := fmt.Sprintf("p%d", i)
			block := int(inodo.I_block[i])
			inode := numInodo
			direccion := fmt.Sprintf("Inodo%d:%s -> Bloque%d:b%d", inode, port, block, block)
			Ap.Direccion = append(Ap.Direccion, direccion)
		}

		if i >= 13 {
			retorno += fmt.Sprintf(`
			<TR><TD>AI %d</TD><TD PORT="p%d">%d</TD></TR>
            `, i, i, inodo.I_block[i])
		} else {
			retorno += fmt.Sprintf(`
			<TR><TD >AD %d</TD><TD PORT="p%d">%d</TD></TR>
            `, i, i, inodo.I_block[i])
		}
	}

	retorno += "\n\t</TABLE>>"
	return retorno
}

func GenerateBloqueCarpetas(bloque Structs.BloquesCarpetas, Inodo int, Sbloque string, NBloque int) string {
	retorno := fmt.Sprintf(`
	<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" BGCOLOR="orange">
		<TR><TD COLSPAN="2" PORT="b%d"><B>Bloque %d</B></TD></TR>`, NBloque, NBloque)

	// Un for que me imprima el contenido de la carpeta
	for i := 0; i < 4; i++ {
		if bloque.B_content[i].B_inodo != -1 {
			if bloque.B_content[i].B_inodo != 0 {

				if i < 4 {
					port := fmt.Sprintf("p%d", i)
					direccion := fmt.Sprintf(" %s:%s:e -> Inodo%d:i%d:w ", Sbloque, port, int(bloque.B_content[i].B_inodo), int(bloque.B_content[i].B_inodo))
					Ap.Direccion = append(Ap.Direccion, direccion)
				} else {
					port := fmt.Sprintf("p%d", i)
					direccion := fmt.Sprintf(" %s:%s -> Inodo%d:i%d ", Sbloque, port, int(bloque.B_content[i].B_inodo), int(bloque.B_content[i].B_inodo))
					Ap.Direccion = append(Ap.Direccion, direccion)
				}
			}

			nombre := ""
			for a := 0; a < len(bloque.B_content[i].B_name); a++ {
				if bloque.B_content[i].B_name[a] != 0 {
					nombre += string(bloque.B_content[i].B_name[a])
				}
			}
			retorno += fmt.Sprintf(`
		<TR><TD>%s</TD><TD PORT="p%d">%d</TD></TR>`, nombre, i, bloque.B_content[i].B_inodo)
		} else {
			retorno += fmt.Sprintf(`
		<TR><TD>--</TD><TD PORT="p%d">%d</TD></TR>`, i, bloque.B_content[i].B_inodo)
		}
	}
	retorno += "\n\t</TABLE>>"
	return retorno
}

func GenerateBloqueArchivo(contenido string, Inodo int, Bloque int) string {
	retorno := fmt.Sprintf(`
	<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" BGCOLOR="green">
		<TR><TD COLSPAN="2" PORT="b%d"><B>Bloque %d</B></TD></TR>
		<TR><TD>%s</TD></TR>
	</TABLE>>`, Bloque, Bloque, contenido)

	return retorno
}

func lecturaB(file *os.File, number int) []byte {
	bytes := make([]byte, number) //array de bytes

	_, err := file.Read(bytes) // Leido -> bytes
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func InterfazCarpetaArchivo(id string) ContenidoFront {

	TamBloqueCarpeta := int(unsafe.Sizeof(Structs.BloquesCarpetas{}))
	TamBloqueArchivo := int(unsafe.Sizeof(Structs.BloquesArchivos{}))
	TamInodo := int(unsafe.Sizeof(Structs.Inodos{}))
	contenido := ContenidoFront{}

	path := id[0:1]
	path = path + ".dsk"

	partcion := Comandos.GetMount("REP", id, &path)
	SB := Structs.NewSuperBloque()
	file, err := os.OpenFile(strings.ReplaceAll(path, "\"", ""), os.O_WRONLY, os.ModeAppend)
	file, err = os.Open(strings.ReplaceAll(path, "\"", ""))
	if err != nil {
		Comandos.Error("REP", "No se ha encontrado el disco.")
		return contenido
	}

	file.Seek(partcion.Part_start, 0)
	data := lecturaB(file, int(unsafe.Sizeof(Structs.SuperBloque{})))
	buffer := bytes.NewBuffer(data)
	err_ := binary.Read(buffer, binary.BigEndian, &SB)
	if err_ != nil {
		Comandos.Error("REP", "Error al leer el archivo")
		return contenido
	}

	MitadBA := (partcion.Part_size - SB.S_block_start) / 2
	MitadBA = MitadBA + SB.S_block_start

	inode := Structs.NewInodos()
	PunteroInodos := SB.S_inode_start
	PunteroBloquesCarpetas := SB.S_block_start
	PunteroBloquesArchivos := MitadBA
	//CantidadBloques := 0

	CantidadBloquesCarpetas := 0
	CantidadBloquesArchivos := 0

	for {
		bc := Structs.NewBloquesCarpetas()
		file.Seek(PunteroBloquesCarpetas, 0)
		data = lecturaB(file, TamBloqueCarpeta)
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &bc)
		if err_ != nil {
			Comandos.Error("MkDir", "Error al leer el archivo")
			return contenido
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

	for {
		var fb Structs.BloquesArchivos
		file.Seek(PunteroBloquesArchivos, 0)
		data = lecturaB(file, TamBloqueArchivo)
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &fb)
		if err_ != nil {
			Comandos.Error("REP", "Error al leer el archivo")
			return contenido
		}
		Contenido := fb.B_content[0]
		if Contenido != 255 {
			CantidadBloquesArchivos++
		} else {
			break
		}
		PunteroBloquesArchivos += int64(TamBloqueArchivo)
	}

	BCN := 0
	BA := 0
	CantidadArchivos := 0

	for i := 0; i < int(SB.S_inodes_count); i++ {
		Carpeta := true
		file.Seek(PunteroInodos, 0)
		data = lecturaB(file, TamInodo)
		buffer = bytes.NewBuffer(data)
		err_ = binary.Read(buffer, binary.BigEndian, &inode)
		if err_ != nil {
			Comandos.Error("REP", "Error al leer el archivo")
			return contenido
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
					Apunta := int(inode.I_block[j])
					if Carpeta == true {

						bc := Structs.NewBloquesCarpetas()

						if Apunta == 0 {
							PunteroBloquesCarpetas = SB.S_block_start + (int64(Apunta) * int64(TamBloqueCarpeta))
						} else {
							resut := int64(Apunta - 2)
							PunteroBloquesCarpetas = SB.S_block_start + resut*int64(TamBloqueCarpeta)
						}

						file.Seek(PunteroBloquesCarpetas, 0)
						data = lecturaB(file, TamBloqueCarpeta)
						buffer = bytes.NewBuffer(data)
						err_ = binary.Read(buffer, binary.BigEndian, &bc)
						if err_ != nil {

							Comandos.Error("REP", "Error al leer el archivo")
							return contenido
						}
						for i := 0; i < 4; i++ {
							if bc.B_content[i].B_inodo != -1 {

								nombre := ""
								for a := 0; a < len(bc.B_content[i].B_name); a++ {
									if bc.B_content[i].B_name[a] != 0 {
										nombre += string(bc.B_content[i].B_name[a])
									}
								}
								if nombre[0] != '.' {
									nuevaCarpeta := CarpetaFront{
										NombreCarpeta: nombre,
									}
									contenido.Carpetas = append(contenido.Carpetas, nuevaCarpeta)
								}

							}
						}
						BCN++

					} else {
						if Apunta < 3 {
							PunteroBloquesArchivos = MitadBA + (int64(Apunta-1) * int64(TamBloqueArchivo))
						} else {
							PunteroBloquesArchivos = MitadBA + (int64(Apunta-CantidadBloquesCarpetas) * int64(TamBloqueArchivo))
						}
						CantidadArchivos += 1
						var fb Structs.BloquesArchivos
						file.Seek(PunteroBloquesArchivos, 0)
						data = lecturaB(file, TamBloqueArchivo)
						buffer = bytes.NewBuffer(data)
						err_ = binary.Read(buffer, binary.BigEndian, &fb)
						if err_ != nil {
							Comandos.Error("REP", "Error al leer el archivo")
							return contenido
						}

						txt := ""
						for i := 0; i < len(fb.B_content); i++ {
							if fb.B_content[i] != 0 {
								txt += string(fb.B_content[i])
							}
							if len(txt) == 64 {
								break
							}
						}
						nuevoArchivo := ArchivoFront{
							NumArchivo: CantidadArchivos,
							Contenido:  txt,
						}
						contenido.Archivos = append(contenido.Archivos, nuevoArchivo)
						BA++
					}

				}
			}
		} else {
			break
		}
		PunteroInodos += int64(TamInodo)
	}

	return contenido
}
