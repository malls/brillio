import { formatMonogram } from '../../format/listing';
import styles from './Monogram.module.css';

export interface MonogramProps {
  /** The listing's city; the block shows its first two letters, e.g. 'Springfield' -> 'Sp'. */
  city: string;
}

// BRANDING.md section 5.5: the image slot's stand-in, a stone block with the city's initials. It
// is texture, not an identifier, and the city is already in the card's meta line, so it is hidden
// from assistive technology.
export function Monogram({ city }: MonogramProps) {
  return (
    <div className={styles.monogram} aria-hidden="true">
      {formatMonogram(city)}
    </div>
  );
}
