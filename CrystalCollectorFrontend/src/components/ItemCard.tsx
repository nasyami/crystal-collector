import React from 'react';

interface ItemCardProps {
  name: string;
  color: string;
  price_cents?: number;
  owned: boolean;
  equipped: boolean;
  onBuy: () => void;
  onApply: () => void;
}

const ItemCard: React.FC<ItemCardProps> = ({
  name,
  color,
  price_cents,
  owned,
  equipped,
  onBuy,
  onApply,
}) => {
  let priceDisplay = 'N/A';
  if (typeof price_cents === 'number' && !isNaN(price_cents)) {
    priceDisplay = `$${(price_cents / 100).toFixed(2)}`;
  }

  return (
    <div style={{ border: '1px solid #333', borderRadius: 8, padding: 16, width: 160, background: '#222' }}>
      <div style={{ width: '100%', height: 56, background: color, borderRadius: 6, marginBottom: 12 }} />
      <h4 style={{ margin: 0 }}>{name}</h4>
      <p style={{ margin: '8px 0 12px 0', color: '#ccc' }}>Price: {priceDisplay}</p>
      {owned ? (
        <button
          type="button"
          disabled={equipped}
          onClick={onApply}
          style={{
            width: '100%',
            padding: '8px 12px',
            borderRadius: 6,
            border: '1px solid #555',
            background: equipped ? '#333' : '#fff',
            color: equipped ? '#aaa' : '#181a20',
            cursor: equipped ? 'default' : 'pointer',
          }}
        >
          {equipped ? 'Applied' : 'Apply'}
        </button>
      ) : (
        <button
          type="button"
          onClick={onBuy}
          style={{
            width: '100%',
            padding: '8px 12px',
            borderRadius: 6,
            border: '1px solid #555',
            background: color,
            color: '#181a20',
            cursor: 'pointer',
          }}
        >
          Buy
        </button>
      )}
    </div>
  );
};

export default ItemCard;
