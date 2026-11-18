import { setupServer } from 'msw/node';

/** Server MSW bersama. Tiap test menambahkan handler sendiri dengan server.use(). */
export const server = setupServer();
