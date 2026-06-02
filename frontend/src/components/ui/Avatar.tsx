interface AvatarProps {
  memberId?: string;
  name?: string;
  size?: number;
  members?: Record<string, { first: string; last: string }>;
}

export function Avatar({ memberId, name, size = 38, members }: AvatarProps) {
  let initials = '?';
  if (memberId && members?.[memberId]) {
    const m = members[memberId];
    initials = (m.first?.[0] ?? '') + (m.last?.[0] ?? '');
  } else if (name) {
    initials = name
      .split(' ')
      .map((w) => w[0])
      .slice(0, 2)
      .join('');
  }

  return (
    <div
      className="sm-avatar"
      style={{ width: size, height: size, fontSize: size * 0.4 }}
    >
      {initials}
    </div>
  );
}
