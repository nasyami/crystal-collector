
import React, { useState, useEffect, useCallback } from 'react';
import Player from '../components/Player';
import Navbar from '../components/Navbar';

const GAME_WIDTH = 600;
const GAME_HEIGHT = 400;
const PLAYER_STEP = 20;

interface GameProps {
  ownedColors: string[];
  equippedColor: string;
  setEquippedColor: (color: string) => void;
}

const Game: React.FC<GameProps> = ({ ownedColors, equippedColor, setEquippedColor }) => {
  const [position, setPosition] = useState({ x: 100, y: 100 });

  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    setPosition(pos => {
      let { x, y } = pos;
      if (e.key === 'ArrowUp') y = Math.max(0, y - PLAYER_STEP);
      if (e.key === 'ArrowDown') y = Math.min(GAME_HEIGHT - 40, y + PLAYER_STEP);
      if (e.key === 'ArrowLeft') x = Math.max(0, x - PLAYER_STEP);
      if (e.key === 'ArrowRight') x = Math.min(GAME_WIDTH - 40, x + PLAYER_STEP);
      return { x, y };
    });
  }, []);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return (
    <div style={{ minHeight: '100vh', background: '#181a20', color: '#fff' }}>
      <Navbar />
      <h2 style={{ textAlign: 'center', marginTop: 24 }}>Game</h2>
      <div style={{ display: 'flex', justifyContent: 'center', gap: 32, margin: '24px 0' }}>
        <div>
          <div style={{ position: 'relative', width: GAME_WIDTH, height: GAME_HEIGHT, border: '2px solid #333', background: '#222', borderRadius: 12 }}>
            <Player x={position.x} y={position.y} color={equippedColor} size={40} />
          </div>
        </div>
        <div style={{ minWidth: 220 }}>
          <h4>Owned Colors</h4>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', marginBottom: 16 }}>
            {ownedColors.map(color => (
              <button
                key={color}
                style={{
                  width: 32,
                  height: 32,
                  background: color,
                  border: equippedColor === color ? '3px solid #fff' : '2px solid #444',
                  borderRadius: 6,
                  cursor: 'pointer',
                  outline: 'none',
                }}
                title={color}
                onClick={() => setEquippedColor(color)}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Game;
