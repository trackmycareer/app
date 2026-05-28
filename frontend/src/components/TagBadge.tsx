import type { Tag } from "@/types";

interface TagBadgeProps {
  tag: Tag;
}

export function TagBadge({ tag }: TagBadgeProps) {
  return (
    <span
      className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
      style={{
        backgroundColor: tag.colour + "20",
        color: tag.colour,
      }}
    >
      {tag.name}
    </span>
  );
}
