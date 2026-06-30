import { setupWorker } from 'msw/browser';
import { handlers } from './handlers';

export const worker = setupWorker(...handlers);

export async function startMocks(): Promise<void> {
  await worker.start({
    onUnhandledRequest: 'bypass', // let real assets / fonts through
    quiet: true,
  });
  // eslint-disable-next-line no-console
  console.info('[mocks] MSW enabled — serving the mock BFF /api/v1');
}
