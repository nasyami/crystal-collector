import React from 'react';


interface PlayerProps {
  x: number;
  y: number;
  color?: string;
  size?: number;
}

const Player: React.FC<PlayerProps> = ({ x, y, color = '#3498db', size = 40 }) => {
  return (
    <div
      style={{
        position: 'absolute',
        left: x,
        top: y,
        width: size,
        height: size,
        background: color,
        borderRadius: 8,
        transition: 'left 0.1s, top 0.1s, width 0.2s, height 0.2s, background 0.2s',
        boxShadow: '0 2px 8px #0004',
      }}
    />
  );
};

export default Player;
