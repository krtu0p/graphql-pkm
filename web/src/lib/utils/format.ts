export function formatDateTime(isoString: string): string {
  if (!isoString) return 'Unknown date';
  const date = new Date(isoString);
  if (Number.isNaN(date.getTime())) {
    return 'Unknown date';
  }
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(date);
}

export function formatRelativeDate(isoString: string): string {
  if (!isoString) return 'Unknown date';
  const date = new Date(isoString);
  const timestamp = date.getTime();
  if (Number.isNaN(timestamp)) {
    return 'Unknown date';
  }

  const diffSeconds = Math.floor((Date.now() - timestamp) / 1000);
  if (diffSeconds < 45) return 'Just now';
  const diffMinutes = Math.floor(diffSeconds / 60);
  if (diffMinutes < 60) return `${diffMinutes}m ago`;
  const diffHours = Math.floor(diffMinutes / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 7) return `${diffDays}d ago`;

  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  }).format(date);
}

export function createExcerpt(content: string, maxLength = 150): string {
  const normalized = content.replace(/\s+/g, ' ').trim();
  if (!normalized) {
    return 'No content written yet.';
  }
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return `${normalized.slice(0, maxLength).trimEnd()}…`;
}

export function normalizeTag(raw: string): string {
  return raw.trim().toLowerCase().replace(/^#+/, '');
}
