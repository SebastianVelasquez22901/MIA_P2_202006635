import React, { useState, useEffect } from 'react';
import './App.css'
import NavBar from './components/NavBar';
import axios from 'axios';
import myImage from './img/disco.png';
import particionImage from './img/particion.png';
import anonymus from './img/user.webp';
import carpetaimg from './img/carpeta.png';
import SVG from './img/svg.png';
import txt from './img/txt.png';
import Swal from 'sweetalert2';
import Button from 'react-bootstrap/Button';
import Modal from 'react-bootstrap/Modal';
import { FaArrowCircleLeft } from "react-icons/fa";
import { Graphviz } from 'graphviz-react';
// or via CommonJS


function App() {
  
  const [selectedOption, setSelectedOption] = useState('Pantalla 1');
  const [chatLines, setChatLines] = useState([]);
  const [inputText, setInputText] = useState('');
  const [discosActivos, setDiscosActivos] = useState(0);
  const [mostrarParticiones, setmostrarParticiones] = useState(false);
  const [particiones, setParticiones] = useState([]);
  const [DiscoSeleccionado, setDiscoSeleccionado] = useState('');
  const [ParticionSeleccionada, setParticionSeleccionada] = useState(0);
  const [userValue, setUserValue] = useState('');
  const [passValue, setPassValue] = useState('');
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [fileContent, setFileContent] = useState('');
  const [Title, setTitle] = useState('');
  const [dotString, setDotString] = useState('');
  const [modalShow, setModalShow] = React.useState(false);
  const [reporteShow, setReporteShow] = React.useState(false);
  const [carpetas, setCarpetas] = useState([]);
  const [archivos, setArchivos] = useState([]);
  const [reportes, setReportes] = useState([]);
  let arrayArchivos = [];


  const LlamarReportes = () => {
    axios.get('http://localhost:3000/reportes').then(function (response) {
      setReportes(response.data.ListaReportes);
      console.log(response.data.ListaReportes);
    })
    .catch(function (error) {
      console.log(error);
    });
  }

  const handleNavClick = (option) => {
    setSelectedOption(option);
    if (option === "Pantalla 3") {
      LlamarReportes();
    }
  }

  function MyVerticallyCenteredModal(props) {
    return (
      <Modal
        {...props}
        size="lg"
        aria-labelledby="contained-modal-title-vcenter"
        centered
      >
        <Modal.Header style={{position: 'relative'}}>
        <Modal.Title id="contained-modal-title-vcenter">
          {Title}
        </Modal.Title>
        <Button 
          onClick={props.onHide} 
          style={{backgroundColor: 'transparent', color: 'black', border: 'none', position: 'absolute', top: '10px', right: '10px'}}
        >
          X
        </Button>
        </Modal.Header>
        <Modal.Body>
          <p>
            {fileContent}
          </p>
        </Modal.Body>
      </Modal>
    );
  }

  function VentanaReport(props) {
  
    return (
      <Modal
        {...props}
        size="lg"
        aria-labelledby="contained-modal-title-vcenter"
        centered
      >
        <Modal.Header style={{position: 'relative'}}>
          <Modal.Title id="contained-modal-title-vcenter">
            {Title}
          </Modal.Title>
          <Button 
            onClick={props.onHide} 
            style={{backgroundColor: 'transparent', color: 'black', border: 'none', position: 'absolute', top: '10px', right: '10px'}}
          >
            X
          </Button>
        </Modal.Header>
        <Modal.Body style={{display: 'flex', justifyContent: 'center', alignItems: 'center'}}>
            <Graphviz dot={dotString} />
        </Modal.Body>
      </Modal>
    );
  }

  const CarpetasArchivos = () => {
    axios.get('http://localhost:3000/carpetaArchivos').then(function (response) {    
    setCarpetas(response.data.Carpetas);
    setArchivos(response.data.Archivos);
    
    })
    .catch(function (error) {
      console.log(error);
    });
  }

  const Login = (event) => {
    event.preventDefault();
    const idValue = DiscoSeleccionado[0] + ParticionSeleccionada + '65';
    
    axios.post('http://localhost:3000/login', {
      id: idValue,
      user: userValue,
      pass: passValue
    })
    .then(response => {
      if (response.status === 200) {
        setIsLoggedIn(true);
        Swal.fire({
          title: "هل تريد الاستمرار؟",
          icon: "question",
          iconHtml: "؟",
          confirmButtonText: "نعم",
          cancelButtonText: "لا",
          showCancelButton: true,
          showCloseButton: true,
          timer: 2000,
          timerProgressBar: true,
          didOpen: () => {
            Swal.showLoading();
          },
          willClose: () => {
            Swal.fire({
              title: "Tranquilo, nadie te esta hackeando",
              text: "Tu sesión ha sido iniciada correctamente",
              icon: "success"
            });
            CarpetasArchivos();
            setSelectedOption('Pantalla 5');
          }
        });
      }
    })
    .catch(error => {
      Swal.fire({
        icon: "error",
        title: "Parece que hubo un error",
        text: "Usuario o contraseña incorrectos"
      });
    });
  }

  const fetchNumber = () => {
    axios.get('http://localhost:3000/verficadorDiscos')
      .then(function (response) {
        setDiscosActivos(parseInt(response.data.number));
      })
      .catch(function (error) {
        console.log(error);
      });
  } 

  const handleSendClick = () => {
    setChatLines([...chatLines, inputText]);
    
    // Hacer una solicitud POST a tu backend
    axios.post('http://localhost:3000/analizador', {
      comandos: [inputText]
    })
    .then(function (response) {
      console.log(response);
    })
    .catch(function (error) {
      console.log(error);
    });
    setInputText('');
  }

  const getParticiones = (letra) => {
    
    setDiscoSeleccionado(letra);
    axios.post('http://localhost:3000/getParticiones', {
      NombreDisco: letra
    })
    .then(function (response) {
      setParticiones(response.data);
    })
    .catch(function (error) {
      console.log(error);
    });
  }

  const handleLogout = () => {
    setIsLoggedIn(false);
    axios.get('http://localhost:3000/logout')
      .then(response => {
        if (response.status === 200) {
          setIsLoggedIn(false);
          Swal.fire(
            '¡Éxito!',
            'Has cerrado sesión correctamente.',
            'success'
          );
        }
      })
      .catch(error => {
        if (error.response && error.response.status === 400) {
          Swal.fire(
            'Error',
            'Hubo un problema al intentar cerrar la sesión.',
            'error'
          );
        }
      });
  };

  const handleButtonClick = (buttonType, esArchivo) => {
    // Buscar en arrayArchivos

    if (!esArchivo) {
      setSelectedOption('Pantalla 6');
    } else {
      arrayArchivos.find((archivo, index) => {
        if (archivo.NombreCarpeta === buttonType) {
          // Guardar la posición
          const posicion = index + 1;
  
          if (buttonType === 'users.txt'){
            const posicion = index + 1;
            const archivoCorrespondiente = archivos.find(archivo => archivo.NumArchivo === posicion);
            if (archivoCorrespondiente.Contenido.length > 60) {
              setTitle(buttonType);
              setFileContent('');
              setFileContent(archivos[0].Contenido + archivos[1 ].Contenido);
              setModalShow(true);
            }
          } else {
            const posicion = index + 2;
            const archivoCorrespondiente = archivos.find(archivo => archivo.NumArchivo === posicion);
            if (archivoCorrespondiente) {
    
              // Hacer console.log del Contenido
              setTitle(buttonType);
              setFileContent(archivoCorrespondiente.Contenido);
              setModalShow(true);
            }
          }
  
          // Buscar en archivos
          
    
          return true;
        }
    
        return false;
      });
    }

    
  };

  const generarVentanaReporte = (reporte) => {
    setReporteShow(true);
    setTitle(reporte.NombreReporte);
    setDotString(reporte.DotString);

  }

  useEffect(() => {
    if (selectedOption === 'Pantalla 2') {
      fetchNumber();
    }
  }, [selectedOption, carpetas, archivos, reportes]);

/*<button 
  style={{display: 'flex', flexDirection: 'column', alignItems: 'center', backgroundColor: '#404040', width: '150px', height: '80px'}}
  onClick={() => handleButtonClick('carpeta')}
>
  <img src={carpeta} alt="Carpeta" style={{width: '100%', height: '100%', objectFit: 'contain'}} />
  <br />
  Carpeta
</button>*/

  return (
    <>
      <NavBar onNavClick={handleNavClick} isLoggedIn={isLoggedIn} onLogout={handleLogout} />
      <div className="gridContainer">
        {selectedOption === 'Pantalla 1' && <div className="gridItem">
          <div className="chatContainer">
            <br />
            <textarea className="chatBox" value={chatLines.join('\n')} readOnly />
            <div className="inputContainer">
              <input type="text" value={inputText} onChange={(e) => setInputText(e.target.value)} />
              <button onClick={handleSendClick}>Enviar</button>
            </div>
          </div>
        </div>}
        {selectedOption === 'Pantalla 2' && <div className="gridItem" style={{width: '100%', height: '90vh'}}>
          <div className="chatContainer">
            <div style={{width: '100%', height: '100%', backgroundColor: '#404040'}}>
              <div style={{display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gridGap: '18px'}}>
            
            {!mostrarParticiones ? Array.from({length: discosActivos}).map((_, index) => {
            const letter = String.fromCharCode(65 + index);
            return (
              <button 
                key={index} 
                style={{display: 'flex', flexDirection: 'column', alignItems: 'center'}}
                onClick={() => {
                  setmostrarParticiones(true);
                  getParticiones(letter+".dsk");
                }}
              >
                <img src={myImage} alt="My Button" style={{width: '100px', height: '100px'}} />
                <br />
                {letter}.dsk
              </button>
            );
          }) : (
            particiones.map((particion, index) => (
              <button 
                key={index} 
                style={{display: 'flex', flexDirection: 'column', alignItems: 'center'}}
                onClick={() => {
                  setParticionSeleccionada(index+1);
                  setSelectedOption('Pantalla 4');
                }}
              >
                <img src={particionImage} alt="Partición" style={{width: '50px', height: '50px'}} />
                <div>{particion.name ? particion.name : "No montada"}</div>
              </button>
            ))
          )}
          </div>
            
            </div>
          </div>
        </div>}
        {selectedOption === 'Pantalla 3' && <div className="gridItem">
        
            <div className="chatContainer">
              <div style={{width: '100%', height: '100%', backgroundColor: '#404040'}}>
              <div style={{display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gridGap: '18px'}}>
              {reportes.map((reporte, index) => {
              // Genera un color aleatorio en formato hexadecimal
              const randomColor = '#' + Math.floor(Math.random()*16777215).toString(16);

              return (
                <button 
                  key={index} 
                  style={{
                    display: 'flex', 
                    flexDirection: 'column', 
                    alignItems: 'center',
                    backgroundColor: randomColor,
                  }}
                  onClick={() => {
                    generarVentanaReporte(reporte);
                  }}
                >
                  <img src={SVG} alt="Reporte" style={{width: '50px', height: '50px'}} />
                  <div>{reporte.NombreReporte ? reporte.NombreReporte : "No disponible"}</div>
                </button>
              );
            })}
                </div>
              </div>
            </div>
        
        </div>}
        {selectedOption === 'Pantalla 4' && <div className="gridItem" style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'flex-start', 
          height: '100vh' 
        }}>
          <form className="neon-form" style={{
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
            alignItems: 'center',
            gap: '10px',
            padding: '20px',
            backgroundColor: '#333',
            borderRadius: '5px',
            marginTop: '200px'
          }}>
            <img src={anonymus} alt="Imagen circular" style={{width: '100px', height: '100px', borderRadius: '50%', marginBottom: '20px'}} />
            <input type="text" placeholder="Usuario" style={{
              padding: '10px', 
              borderRadius: '5px', 
              color: '#0ff', 
              textShadow: '0 0 3px #0ff, 0 0 6px #0ff'
            }}
            onChange={e => setUserValue(e.target.value)}
            />
            <input type="password" placeholder="Contraseña" style={{
              padding: '10px', 
              borderRadius: '5px', 
              color: '#0ff', 
              textShadow: '0 0 3px #0ff, 0 0 6px #0ff'
            }}
            onChange={e => setPassValue(e.target.value)}
            />
      <button className="neon-button" onClick={(event)=>Login(event)}>
        <span></span>
        <span></span>
        <span></span>
        <span></span>
        Iniciar sesión
      </button>
          </form>
        </div>}
        {selectedOption === 'Pantalla 5' && 
          <div className="gridItem">
            <div className="chatContainer">
              <div style={{width: '100%', height: '100%', backgroundColor: '#404040'}}>
                <div style={{display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gridGap: '4px'}}>
                  
                  
                  {carpetas.map((carpeta, index) => {
                    const esArchivo = carpeta.NombreCarpeta.includes('.');
                    const imagen = esArchivo ? txt : carpetaimg; // Asegúrate de tener una imagen para 'carpeta'
                    if (esArchivo) {
                      arrayArchivos.push({NombreCarpeta: carpeta.NombreCarpeta});
                    }

                    return (
                      <button 
                        key={index}
                        style={{display: 'flex', flexDirection: 'column', alignItems: 'center', backgroundColor: '#404040', width: '150px', height: '80px'}}
                        onClick={() => handleButtonClick(carpeta.NombreCarpeta, esArchivo)}
                      >
                        <img src={imagen} alt={carpeta.NombreCarpeta} style={{width: '100%', height: '100%', objectFit: 'contain'}} />
                        <br />
                        {carpeta.NombreCarpeta}
                      </button>
                    );  
                  })}
                </div>
              </div>
            </div>
          </div>
        }
        {selectedOption === 'Pantalla 6' && 
          <div className="gridItem">
            <div className="chatContainer">
            <div style={{width: '100%', height: '100%', backgroundColor: '#404040'}}>
            <button 
                        style={{display: 'flex', flexDirection: 'column', alignItems: 'center', backgroundColor: '#404040', width: '150px', height: '80px'}}
                        onClick={() => setSelectedOption('Pantalla 5')}
                      >
                        
                        <br />
                        <FaArrowCircleLeft />
                      </button>
            </div>
            </div>
          </div>}
        
      </div>
      <MyVerticallyCenteredModal
        show={modalShow}
        onHide={() => setModalShow(false)}
      />
      <VentanaReport
        show={reporteShow}
        onHide={() => setReporteShow(false)}
      />
    </>
  )
}

export default App