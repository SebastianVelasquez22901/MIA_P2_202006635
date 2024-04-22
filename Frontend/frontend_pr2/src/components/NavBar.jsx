import React from 'react';

const NavBar = ({ onNavClick }) => {
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
            </ul>
        </nav>
    );
}

export default NavBar;