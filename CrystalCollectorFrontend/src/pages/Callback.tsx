import { useEffect, useState } from 'react';

const Callback = () => {
  const [error, setError] = useState('');

  useEffect(() => {
    console.log(window.location.href);
    const params = new URLSearchParams(window.location.search);
    const token = params.get('token');

    if (token) {
      localStorage.setItem('accessToken', token);
      window.history.replaceState({}, document.title, window.location.pathname);
      window.location.replace('/shop');
      return;
    }

    setError('Login failed: token is missing.');
  }, []);

  return (
    <div>
      <h2>Login Error</h2>
      <p>{error}</p>
    </div>
  );
};

export default Callback;
