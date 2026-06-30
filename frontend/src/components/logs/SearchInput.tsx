import { Search } from 'lucide-react';
import { Input } from '@/components/ui/Input';
import { useUiStore } from '@/store/uiStore';

export function SearchInput() {
  const search = useUiStore((s) => s.search);
  const setSearch = useUiStore((s) => s.setSearch);

  return (
    <Input
      leadingIcon={<Search size={16} />}
      value={search}
      onChange={(e) => setSearch(e.target.value)}
      placeholder="Search logs..."
      aria-label="Search logs"
      type="search"
    />
  );
}
