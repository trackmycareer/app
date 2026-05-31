export function ConfettiPiece({ index }: { index: number }) {
  const left = Math.random() * 100;
  const delay = Math.random() * 0.5;
  const hue = Math.floor(Math.random() * 360);
  const size = 6 + Math.random() * 6;

  return (
    <div
      className="confetti-piece absolute top-0"
      style={{
        left: `${left}%`,
        width: `${size}px`,
        height: `${size * 0.6}px`,
        backgroundColor: `hsl(${hue}, 80%, 60%)`,
        animationDelay: `${delay}s`,
        borderRadius: "2px",
      }}
      aria-hidden="true"
      key={index}
    />
  );
}
