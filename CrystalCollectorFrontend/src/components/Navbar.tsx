import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { getCurrentUserFromToken } from '../auth';

interface NavbarProps {
  onLogout?: () => void;
}

const Navbar = ({ onLogout }: NavbarProps) => {
  const [currentUser, setCurrentUser] = useState(getCurrentUserFromToken);
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem('accessToken');
    setCurrentUser(null);
    onLogout?.();
    navigate('/');
  };

  return (
    <nav style={{ padding: '10px', borderBottom: '1px solid #ccc', marginBottom: 20, display: 'flex', alignItems: 'center', gap: 16 }}>
      <Link to="/" style={{ marginRight: 16 }}>Home</Link>
      <Link to="/game" style={{ marginRight: 16 }}>Game</Link>
      <Link to="/shop">Shop</Link>
      <span style={{ marginLeft: 'auto', fontSize: 14, color: currentUser ? '#69db7c' : '#aaa' }}>
        {currentUser ? `Logged in as: ${currentUser.email}` : 'Not logged in'}
      </span>
      {currentUser && (
        <button
          onClick={handleLogout}
          style={{ background: '#ff6b6b', color: '#fff', border: 'none', borderRadius: 6, padding: '4px 12px', cursor: 'pointer' }}
        >
          Logout
        </button>
      )}
    </nav>
  );
};

export default Navbar;
