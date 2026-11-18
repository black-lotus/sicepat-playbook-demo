import { createBrowserRouter, Navigate, type RouteObject } from 'react-router-dom';

import { BranchSummaryPage } from '../pages/branch-summary/BranchSummaryPage';
import { CourierPerformancePage } from '../pages/courier-performance/CourierPerformancePage';
import { FailedDeliveryPage } from '../pages/failed-delivery/FailedDeliveryPage';
import { Layout } from './Layout';
import { RouteError } from './RouteError';

export const routes: RouteObject[] = [
  {
    element: <Layout />,
    errorElement: <RouteError />,
    children: [
      { index: true, element: <Navigate to="/courier-performance" replace /> },
      { path: 'courier-performance', element: <CourierPerformancePage /> },
      { path: 'branch-summary', element: <BranchSummaryPage /> },
      { path: 'failed-delivery', element: <FailedDeliveryPage /> },
    ],
  },
];

export const createAppRouter = () => createBrowserRouter(routes);
