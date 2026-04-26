import { Link } from 'react-router-dom';
import LoginButton from '../components/LoginButton';
import Navbar from '../components/Navbar';

const Home = () => (
  <div>
    <Navbar />
    <h2>Welcome to Crystal Collector!</h2>
    <div style={{ marginTop: 16 }}>
      <LoginButton />
    </div>
    <div style={{ marginTop: 24 }}>
      <Link to="/game" style={{ marginRight: 16 }}>Go to Game</Link>
      <Link to="/shop">Go to Shop</Link>
    </div>
  </div>
);

export default Home;
