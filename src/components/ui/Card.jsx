/**
 * Card Component System (MAT-154 Phase 2)
 *
 * Shared card container with header/body/footer structure.
 * Replaces duplicated card implementations in BoxerCard, RankCard, FightDetails.
 *
 * Usage:
 *   <Card variant="default" clickable onClick={handleClick}>
 *     <CardHeader>...</CardHeader>
 *     <CardBody>...</CardBody>
 *     <CardFooter>...</CardFooter>
 *   </Card>
 *
 * Variants: 'default' (navy), 'light' (premium), 'winner' (gold accent)
 *
 * @param {Object} children - Card content
 * @param {'default'|'light'|'winner'} variant - Visual variant
 * @param {boolean} clickable - Add hover/active states
 * @param {Function} onClick - Click handler for clickable cards
 * @param {string} className - Additional CSS classes
 */

import './Card.css';

export function Card({ children, variant = 'default', clickable = false, onClick, className = '' }) {
  const cardClasses = [
    'card',
    `card-${variant}`,
    clickable ? 'card-clickable' : '',
    className
  ].filter(Boolean).join(' ');

  return (
    <div className={cardClasses} onClick={onClick && clickable ? onClick : undefined}>
      {children}
    </div>
  );
}

/**
 * CardHeader - Card header section with avatar and title
 *
 * @param {Object} children - Header content
 * @param {string} className - Additional CSS classes
 */
export function CardHeader({ children, className = '' }) {
  return (
    <div className={`card-header ${className}`}>
      {children}
    </div>
  );
}

/**
 * CardBody - Card body section for main content
 *
 * @param {Object} children - Body content
 * @param {string} className - Additional CSS classes
 */
export function CardBody({ children, className = '' }) {
  return (
    <div className={`card-body ${className}`}>
      {children}
    </div>
  );
}

/**
 * CardFooter - Card footer section for metadata/record
 *
 * @param {Object} children - Footer content
 * @param {string} className - Additional CSS classes
 */
export function CardFooter({ children, className = '' }) {
  return (
    <div className={`card-footer ${className}`}>
      {children}
    </div>
  );
}
