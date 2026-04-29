import { useEffect, useMemo, useState } from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import { api } from './api';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081';
import Home from './pages/Home';
import Game from './pages/Game';
import Shop from './pages/Shop';
import Callback from './pages/Callback';

export interface ColorItem {
  id: string;
  name: string;
  description?: string;
  sku?: string;
  price_cents?: number;
}

interface OwnedItemsResponse {
  items: ColorItem[];
}

interface PaymentTokenResponse {
  pay_station_url?: string;
}

export const DEFAULT_COLOR = '#4fd1c5';

export const getItemColor = (item: ColorItem): string => {
  const normalized = item.name.trim().toLowerCase();

  if (normalized.includes('red')) return '#ff6b6b';
  if (normalized.includes('blue')) return '#4dabf7';
  if (normalized.includes('green')) return '#69db7c';
  if (normalized.includes('purple')) return '#b197fc';
  if (normalized.includes('gold')) return '#ffd700';

  return DEFAULT_COLOR;
};

const App = () => {
  const [items, setItems] = useState<ColorItem[]>([]);
  const [ownedItemIds, setOwnedItemIds] = useState<string[]>([]);
  const [equippedColor, setEquippedColor] = useState<string>(DEFAULT_COLOR);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [buyStatus, setBuyStatus] = useState('');

  const refreshOwnedItems = async () => {
    try {
      console.log('[App] calling /v1/me/items');
      const ownedRes = await api.get<OwnedItemsResponse>(`${API_BASE_URL}/v1/me/items`);
      console.log('[App] /v1/me/items response:', ownedRes.data.items);
      setOwnedItemIds((ownedRes.data.items ?? []).map((item) => item.id));
    } catch {
      setOwnedItemIds([]);
    }
  };

  useEffect(() => {
    const loadShop = async () => {
      setLoading(true);
      setError('');

      const token = localStorage.getItem('accessToken');
      console.log('[App] token exists:', !!token);
      if (token) {
        try {
          const decoded = JSON.parse(atob(token.split('.')[1]));
          console.log('[App] decoded user:', decoded);
        } catch {
          console.log('[App] failed to decode token');
        }
      }

      try {
        const itemsRes = await api.get<ColorItem[]>(`${API_BASE_URL}/v1/items`);
        setItems(itemsRes.data ?? []);
      } catch {
        setError('Failed to load shop items');
      }

      await refreshOwnedItems();
      setLoading(false);
    };

    void loadShop();
  }, []);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('status') === 'done') {
      setBuyStatus('Payment successful');
      void refreshOwnedItems();
    }
  }, []);

  const ownedColors = useMemo(
    () =>
      items
        .filter((item) => ownedItemIds.includes(item.id))
        .map((item) => getItemColor(item)),
    [items, ownedItemIds]
  );

  const handleLogout = () => {
    setOwnedItemIds([]);
  };

  const handleBuyItem = async (itemId: string) => {
    try {
      setError('');
      setBuyStatus('Processing payment...');
      const response = await api.post<PaymentTokenResponse>(
        `${API_BASE_URL}/v1/payments/token`,
        { item_id: itemId },
      );
      console.log(response.data);

      if (response.data.pay_station_url) {
        window.location.href = response.data.pay_station_url;
      }

      setOwnedItemIds((current) =>
        current.includes(itemId) ? current : [...current, itemId]
      );
    } catch {
      setBuyStatus('');
      setError('Failed to start checkout');
    }
  };

  const handleApplyItem = (item: ColorItem) => {
    if (ownedItemIds.includes(item.id)) {
      setEquippedColor(getItemColor(item));
    }
  };

  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/callback" element={<Callback />} />
        <Route
          path="/game"
          element={
            <Game
              ownedColors={ownedColors}
              equippedColor={equippedColor}
              setEquippedColor={setEquippedColor}
            />
          }
        />
        <Route
          path="/shop"
          element={
            <Shop
              items={items}
              ownedItemIds={ownedItemIds}
              equippedColor={equippedColor}
              loading={loading}
              error={error}
              buyStatus={buyStatus}
              onBuyItem={handleBuyItem}
              onApplyItem={handleApplyItem}
              onLogout={handleLogout}
            />
          }
        />
      </Routes>
    </Router>
  );
};

export default App;
