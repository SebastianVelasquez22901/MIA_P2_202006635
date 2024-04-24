import React from 'react';

const NavBar = ({ onNavClick, isLoggedIn, onLogout }) => {
    return (
        <nav className="navBar">
            <ul className="navList">
                <li className="navItem">
                    <a href="#" className="navLink" onClick={() => onNavClick('Pantalla 1')}>Pantalla 1</a>
                </li>
                <li className="navItem">
                    <a href="#" className="navLink" onClick={() => onNavClick('Pantalla 2')}>Pantalla 2</a>
                </li>
                <li className="navItem">
                    <a href="#" className="navLink" onClick={() => onNavClick('Pantalla 3')}>Pantalla 3</a>
                </li>
                {isLoggedIn && (
                  <li className="navItem">
                    <button style={{backgroundColor: 'red'}} onClick={onLogout}>
                      Salir sesión
                    </button>
                  </li>
                )}
            </ul>
        </nav>
    );
}

export default NavBar;