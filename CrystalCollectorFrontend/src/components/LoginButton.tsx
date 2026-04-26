import { Widget } from '@xsolla/login-sdk';

const LoginButton = () => {
  const handleLogin = () => {
    try {
      console.log('Login button clicked');
      console.log(
        'VITE_XSOLLA_LOGIN_PROJECT_ID:',
        import.meta.env.VITE_XSOLLA_LOGIN_PROJECT_ID
      );
      console.log('VITE_XSOLLA_CLIENT_ID:', import.meta.env.VITE_XSOLLA_CLIENT_ID);
      console.log(
        'VITE_XSOLLA_REDIRECT_URI:',
        import.meta.env.VITE_XSOLLA_REDIRECT_URI
      );

      const redirectUri = import.meta.env.VITE_XSOLLA_REDIRECT_URI;
      console.log('Final redirectUri:', redirectUri);
      console.log('Creating Xsolla Login Widget');

      const xl = new Widget({
        projectId: import.meta.env.VITE_XSOLLA_LOGIN_PROJECT_ID,
        preferredLocale: 'en',
        clientId: import.meta.env.VITE_XSOLLA_CLIENT_ID,
        responseType: 'token',
        callbackUrl: redirectUri,
        redirectUri: redirectUri,
      });

      const container = document.getElementById('xl_auth');
      if (container) {
        container.style.display = 'block';
      }

      xl.mount('xl_auth');
      console.log('Calling xl.open()');
      xl.open();
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <>
      <button type="button" onClick={handleLogin}>
        Login
      </button>
      <div id="xl_auth" style={{ display: 'none' }} />
    </>
  );
};

export default LoginButton;
