import { LayoutList, RefreshCw, Table } from 'lucide-react';
import { SegmentedControl } from '@/components/ui/SegmentedControl';
import { Button } from '@/components/ui/Button';
import type { LogsView } from '@/store/uiStore';

interface ViewToggleProps {
  view: LogsView;
  onViewChange: (v: LogsView) => void;
  onReload: () => void;
  isFetching: boolean;
}

export function ViewToggle({ view, onViewChange, onReload, isFetching }: ViewToggleProps) {
  return (
    <div className="flex gap-2">
      <SegmentedControl<LogsView>
        aria-label="View"
        options={[
          { value: 'table', label: 'Table', icon: <Table size={15} aria-hidden="true" /> },
          { value: 'cards', label: 'Cards', icon: <LayoutList size={15} aria-hidden="true" /> },
        ]}
        value={view}
        onChange={onViewChange}
      />
      <Button
        variant="outline"
        size="sm"
        onClick={onReload}
        loading={isFetching}
        leftIcon={<RefreshCw size={15} aria-hidden="true" />}
      >
        Reload
      </Button>
    </div>
  );
}
