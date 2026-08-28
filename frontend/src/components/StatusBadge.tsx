interface StatusBadgeProps {
  status: 'available' | 'confirmed' | 'pending' | 'maintenance' | 'completed';
}

const config: Record<StatusBadgeProps['status'], { label: string; dotColor: string; bgColor: string; textColor: string }> = {
  available: {
    label: 'Available',
    dotColor: 'bg-success',
    bgColor: 'bg-blush-light',
    textColor: 'text-plum',
  },
  confirmed: {
    label: 'Confirmed',
    dotColor: 'bg-success',
    bgColor: 'bg-blush-light',
    textColor: 'text-plum',
  },
  pending: {
    label: 'Pending',
    dotColor: 'bg-warning',
    bgColor: 'bg-blush-light',
    textColor: 'text-plum',
  },
  maintenance: {
    label: 'Maintenance',
    dotColor: 'bg-danger',
    bgColor: 'bg-blush-light',
    textColor: 'text-danger',
  },
  completed: {
    label: 'Completed',
    dotColor: 'bg-plum-muted',
    bgColor: 'bg-blush-light',
    textColor: 'text-plum-muted',
  },
};

export default function StatusBadge({ status }: StatusBadgeProps) {
  const c = config[status];
  return (
    <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium ${c.bgColor} ${c.textColor}`}>
      <span className={`w-1.5 h-1.5 rounded-full ${c.dotColor}`} />
      {c.label}
    </span>
  );
}
