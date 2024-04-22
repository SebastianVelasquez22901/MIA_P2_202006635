import React, { useState, useEffect } from 'react';
import './App.css'
import NavBar from './components/NavBar';
import axios from 'axios';
import myImage from './img/disco.png';

function App() {
  const [selectedOption, setSelectedOption] = useState('Pantalla 1');
  const [chatLines, setChatLines] = useState([]);
  const [inputText, setInputText] = useState('');
  const [discosActivos, setDiscosActivos] = useState(0);

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
              <div style={{display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', gridGap: '18px'}}>
              {Array.from({length: discosActivos}).map((_, index) => (
              <button key={index} style={{display: 'flex', flexDirection: 'column', alignItems: 'center'}}>
                <img src={myImage} alt="My Button" style={{width: '100px', height: '100px'}} />
              <br />
              {String.fromCharCode(65 + index)}.dsk</button>
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