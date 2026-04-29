import React, { useEffect } from 'react';
import ItemCard from '../components/ItemCard';
import Navbar from '../components/Navbar';
import { getItemColor, type ColorItem } from '../App';
import { GTAG_EVENTS, fireEvent } from '../types/gtag';

interface ShopProps {
  items: ColorItem[];
  ownedItemIds: string[];
  equippedColor: string;
  loading: boolean;
  error: string;
  buyStatus: string;
  onBuyItem: (itemId: string) => void;
  onApplyItem: (item: ColorItem) => void;
  onLogout: () => void;
}

const Shop: React.FC<ShopProps> = ({
  items,
  ownedItemIds,
  equippedColor,
  loading,
  error,
  buyStatus,
  onBuyItem,
  onApplyItem,
  onLogout,
}) => {
  useEffect(() => {
    if (!loading) {
      fireEvent(GTAG_EVENTS.VIEW_ITEM_LIST, { count: items.length });
    }
    // Only fire when loading changes to false
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading]);
  return (
    <div style={{ minHeight: '100vh', background: '#181a20', color: '#fff' }}>
      <Navbar onLogout={onLogout} />
      <h2 style={{ textAlign: 'center', marginTop: 24 }}>Shop</h2>
      {loading && <p style={{ textAlign: 'center' }}>Loading...</p>}
      {error && <p style={{ color: '#ff6b6b', textAlign: 'center' }}>{error}</p>}
      {buyStatus && <p style={{ color: buyStatus === 'Payment successful' ? '#69db7c' : '#ffd43b', textAlign: 'center' }}>{buyStatus}</p>}
      <div style={{ display: 'flex', flexWrap: 'wrap', justifyContent: 'center', gap: 12, padding: 24 }}>
        {items.map(item => {
          const color = getItemColor(item);

          return (
            <ItemCard
              key={item.id}
              name={item.name}
              color={color}
              price_cents={item.price_cents}
              owned={ownedItemIds.includes(item.id)}
              equipped={equippedColor === color}
              onBuy={() => onBuyItem(item.id)}
              onApply={() => onApplyItem(item)}
            />
          );
        })}
      </div>
    </div>
  );
};

export default Shop;
