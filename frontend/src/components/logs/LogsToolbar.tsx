import { ProviderSelect } from './ProviderSelect';
import { SearchInput } from './SearchInput';

export function LogsToolbar() {
  return (
    <div className="mb-3.5 flex items-center gap-3.5">
      <div className="flex items-center gap-2">
        <span className="text-label text-muted">Provider</span>
        <ProviderSelect />
      </div>
      <div className="flex-1">
        <SearchInput />
      </div>
    </div>
  );
}
