import { Link } from 'react-router-dom';

const Navbar = () => (
  <nav style={{ padding: '10px', borderBottom: '1px solid #ccc', marginBottom: 20 }}>
    <Link to="/" style={{ marginRight: 16 }}>Home</Link>
    <Link to="/game" style={{ marginRight: 16 }}>Game</Link>
    <Link to="/shop">Shop</Link>
  </nav>
);

export default Navbar;
