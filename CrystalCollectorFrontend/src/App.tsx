import { useEffect, useMemo, useState } from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import { api } from './api';
import Home from './pages/Home';
import Game from './pages/Game';
import Shop from './pages/Shop';
import Callback from './pages/Callback';

export interface ColorItem {
  id: string;
  name: string;
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
  if (normalized.includes('yellow')) return '#ffd43b';
  if (normalized.includes('purple')) return '#b197fc';
  if (normalized.includes('orange')) return '#ffa94d';
  if (normalized.includes('pink')) return '#f783ac';

  return DEFAULT_COLOR;
};

const App = () => {
  const [items, setItems] = useState<ColorItem[]>([]);
  const [ownedItemIds, setOwnedItemIds] = useState<string[]>([]);
  const [equippedColor, setEquippedColor] = useState<string>(DEFAULT_COLOR);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const loadShop = async () => {
      setLoading(true);
      setError('');

      try {
        const itemsRes = await api.get<ColorItem[]>(`${API_BASE_URL}/v1/items`);
        setItems(itemsRes.data);
      } catch {
        setError('Failed to load shop items');
      }

      try {
        const ownedRes = await api.get<OwnedItemsResponse>(`${API_BASE_URL}/v1/me/items`);
        setOwnedItemIds(ownedRes.data.items.map((item) => item.id));
      } catch {
        setOwnedItemIds([]);
      } finally {
        setLoading(false);
      }
    };

    void loadShop();
  }, []);

  const ownedColors = useMemo(
    () =>
      items
        .filter((item) => ownedItemIds.includes(item.id))
        .map((item) => getItemColor(item)),
    [items, ownedItemIds]
  );

  const handleBuyItem = async (itemId: string) => {
    try {
      setError('');
      const accessToken = localStorage.getItem('accessToken');
      console.log('accessToken exists:', Boolean(accessToken));

      const { data: response } = await api.post<PaymentTokenResponse>(
        `${API_BASE_URL}/v1/payments/token`,
        {
          item_id: itemId,
        },
      );
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        }
      );
      console.log(response);

      if (response.pay_station_url) {
        window.location.href = response.pay_station_url;
      }

      setOwnedItemIds((current) =>
        current.includes(itemId) ? current : [...current, itemId]
      );
    } catch {
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
              onBuyItem={handleBuyItem}
              onApplyItem={handleApplyItem}
            />
          }
        />
      </Routes>
    </Router>
  );
};

export default App;
