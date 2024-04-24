import React, { useState, useEffect } from 'react';
import './App.css'
import NavBar from './components/NavBar';
import axios from 'axios';
import myImage from './img/disco.png';
import particionImage from './img/particion.png';

function App() {
  const [selectedOption, setSelectedOption] = useState('Pantalla 1');
  const [chatLines, setChatLines] = useState([]);
  const [inputText, setInputText] = useState('');
  const [discosActivos, setDiscosActivos] = useState(0);
  const [mostrarParticiones, setmostrarParticiones] = useState(false);
  const [particiones, setParticiones] = useState([]);
  
  const handleNavClick = (option) => {
    setSelectedOption(option);
  }

  const fetchNumber = () => {
    axios.get('http://localhost:3000/verficadorDiscos')
      .then(function (response) {
        setDiscosActivos(parseInt(response.data.number));
        console.log(response.data.number); // { "number": "3" }
      })
      .catch(function (error) {
        console.log(error);
      });
  } 

  const handleSendClick = () => {
    setChatLines([...chatLines, inputText]);
    
    console.log('Enviando comando: ', inputText);
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
    console.log('Enviando solicitud de particiones para disco: ', letra);
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

  useEffect(() => {
    if (selectedOption === 'Pantalla 2') {
      fetchNumber();
    }
  }, [selectedOption]);



  return (
    <>
      <NavBar onNavClick={handleNavClick} />
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
            }) : particiones.map((particion, index) => (
              <button key={index} style={{display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
                <img src={particionImage} alt="Partición" style={{width: '50px', height: '50px'}} />
                <div>{particion.name ? particion.name : "No montada"}</div>
              </button>
            ))}
          </div>
            
            </div>
          </div>
        </div>}
        {selectedOption === 'Pantalla 3' && <div className="gridItem">Grid for Option 3</div>}
      </div>
    </>
  )
}

export default App