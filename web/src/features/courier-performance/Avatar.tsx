import type { CSSProperties } from 'react';

import { avatarHue, initials } from '../../lib/avatar';

interface AvatarProps {
  courierID: string;
  name: string;
  /** Untuk peringkat 1 sampai 3: avatar memakai warna medali, bukan hue dari courier_id. */
  medal?: 'gold' | 'silver' | 'bronze';
  size?: 'sm' | 'md' | 'lg';
}

/** Avatar lingkaran berisi inisial. Dekoratif: nama kurir selalu ditulis di dekatnya. */
export function Avatar({ courierID, name, medal, size = 'md' }: AvatarProps) {
  const style = { '--hue': avatarHue(courierID) } as CSSProperties;
  const className = `avatar avatar-${size}${medal ? ` avatar-medal avatar-${medal}` : ''}`;
  return (
    <span className={className} style={style} aria-hidden="true">
      {initials(name)}
    </span>
  );
}
