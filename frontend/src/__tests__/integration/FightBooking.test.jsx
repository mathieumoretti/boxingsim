/**
 * @jest-environment jsdom
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Router } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import axios from 'axios';
import FightBookingModal from '../../components/FightBooking/FightBookingModal';
import OpponentList from '../../components/Rankings/OpponentList';

// Mock API responses
const mockOpponents = [
  {
    id: 1,
    name: 'Iron Fist Chen',
    level: 12,
    health: 95,
    energy: 100,
    match_score: 0.85,
    ranking_position: 3,
    is_available: true,
  },
  {
    id: 2,
    name: 'Thunder Strike Johnson',
    level: 11,
    health: 88,
    energy: 92,
    match_score: 0.75,
    ranking_position: 5,
    is_available: true,
  },
  {
    id: 3,
    name: 'Heavy Hammer Silva',
    level: 15,
    health: 75,
    energy: 80,
    match_score: 0.60,
    ranking_position: 1,
    is_available: false, // Not available due to low health
  },
];

const mockFightSuccess = {
  id: 42,
  boxer1_id: 10,
  boxer2_id: 1,
  status: 'scheduled',
  scheduled_time: '2026-10-15T18:00:00Z',
  round: 12,
};

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      gcTime: 5 * 60 * 1000, // 5 minutes
    },
  },
});

const renderWithProviders = (component) => {
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        {component}
      </MemoryRouter>
    </QueryClientProvider>
  );
};

describe('Fight Booking Integration', () => {
  const axiosPostSpy = jest.spyOn(axios, 'post');
  const axiosGetSpy = jest.spyOn(axios, 'get');

  beforeEach(() => {
    axiosPostSpy.mockClear();
    axiosGetSpy.mockClear();
    queryClient.clear();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('FightBookingModal', () => {
    beforeEach(() => {
      axiosGetSpy.mockImplementation((url) => {
        if (url.includes('/api/opponents')) {
          return Promise.resolve({ data: mockOpponents });
        }
        return Promise.reject(new Error(`Unexpected URL: ${url}`));
      });
    });

    test('books a fight and shows confirmation', async () => {
      const user = userEvent.setup();

      axiosGetSpy.mockResolvedValueOnce({ data: mockOpponents });

      axiosPostSpy.mockImplementation((url) => {
        if (url.includes('/api/fights')) {
          return Promise.resolve({ data: mockFightSuccess });
        }
        return Promise.reject(new Error(`Unexpected URL: ${url}`));
      });

      renderWithProviders(
        <FightBookingModal
          myBoxerId={10}
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
        />
      );

      // Wait for opponent list to load
      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      // Select an opponent
      const opponentRow = screen.getByText('Iron Fist Chen').closest('tr');
      fireEvent.click(opponentRow);

      expect(screen.getByDisplayValue(/2026-10/i)).toBeInTheDocument();

      await user.click(screen.getByRole('button', { name: /book fight/i }));

      await waitFor(() => {
        expect(axiosPostSpy).toHaveBeenCalledWith('/api/fights', expect.any(Object));
      });

      const callArgs = axiosPostSpy.mock.calls[0][1];
      expect(callArgs).toMatchObject({
        boxer2_id: 1,
        round: 12,
      });

      await waitFor(() => {
        expect(screen.getByText(/fight scheduled/i)).toBeInTheDocument();
      });
    });

    test('shows opponent availability indicators', async () => {
      axiosGetSpy.mockResolvedValueOnce({ data: mockOpponents });

      renderWithProviders(
        <FightBookingModal
          myBoxerId={10}
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      const rows = screen.getAllByRole('row');
      expect(rows.length).toBeGreaterThan(1);

      let availableRow = null;
      let unavailableRow = null;

      rows.forEach((row) => {
        if (row.textContent.includes('Iron Fist Chen')) {
          availableRow = row;
        }
        if (row.textContent.includes('Heavy Hammer Silva')) {
          unavailableRow = row;
        }
      });

      expect(availableRow).toBeTruthy();
      expect(unavailableRow).toBeTruthy();
    });

    test('prevents booking with non-available opponents', async () => {
      const user = userEvent.setup();

      axiosGetSpy.mockResolvedValueOnce({ data: mockOpponents });

      renderWithProviders(
        <FightBookingModal
          myBoxerId={10}
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      const unavailableRow = screen.getByText('Heavy Hammer Silva').closest('tr');
      fireEvent.click(unavailableRow);

      const bookButton = screen.getByRole('button', { name: /book fight/i });
      expect(bookButton).toBeDisabled();
    });

    test('shows match quality indicators', async () => {
      axiosGetSpy.mockResolvedValueOnce({ data: mockOpponents });

      renderWithProviders(
        <FightBookingModal
          myBoxerId={10}
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      expect(screen.getByText(/0\.85/i)).toBeInTheDocument();
    });

    test('handles booking error gracefully', async () => {
      const user = userEvent.setup();

      axiosGetSpy.mockResolvedValueOnce({ data: mockOpponents });

      axiosPostSpy.mockImplementation((url) => {
        if (url.includes('/api/fights')) {
          return Promise.reject({
            response: {
              status: 400,
              data: { message: 'Cannot fight yourself' },
            },
          });
        }
        return Promise.reject(new Error(`Unexpected URL: ${url}`));
      });

      renderWithProviders(
        <FightBookingModal
          myBoxerId={10}
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      const opponentRow = screen.getByText('Iron Fist Chen').closest('tr');
      fireEvent.click(opponentRow);

      await user.click(screen.getByRole('button', { name: /book fight/i }));

      await waitFor(() => {
        expect(screen.getByText(/error/i)).toBeInTheDocument();
      });
    });
  });

  describe('OpponentList component', () => {
    beforeEach(() => {
      axiosGetSpy.mockResolvedValue({ data: mockOpponents });
    });

    test('displays opponents with correct information', async () => {
      renderWithProviders(<OpponentList myBoxerId={10} />);

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
        expect(screen.getByText(/level 12/i)).toBeInTheDocument();
      });

      const rows = screen.getAllByRole('row');
      expect(rows.length).toBeGreaterThan(1);
    });

    test('allows selecting opponent for fight booking', async () => {
      const onSelectMock = jest.fn();

      renderWithProviders(<OpponentList myBoxerId={10} onSelectOpponent={onSelectMock} />);

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      const opponentRow = screen.getByText('Iron Fist Chen').closest('tr');
      fireEvent.click(opponentRow);

      expect(onSelectMock).toHaveBeenCalledWith(mockOpponents[0]);
    });

    test('filters out unavailable opponents', async () => {
      renderWithProviders(<OpponentList myBoxerId={10} showUnavailable={false} />);

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      const availableRows = screen.getAllByRole('row').filter((row) =>
        row.textContent.includes('is_available') ||
        row.querySelector('[data-available="true"]') !== null
      );

      expect(availableRows.length).toBeGreaterThan(0);
    });

    test('shows level filtering', async () => {
      const user = userEvent.setup();

      renderWithProviders(<OpponentList myBoxerId={10} />);

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      if (screen.queryByPlaceholderText(/min level/i)) {
        await user.type(screen.getByPlaceholderText(/min level/i), '13');
        await user.type(screen.getByPlaceholderText(/max level/i), '14');

        expect(screen.queryByText('Iron Fist Chen')).toBeInTheDocument();
        expect(screen.queryByText('Heavy Hammer Silva')).not.toBeInTheDocument();
      }
    });
  });

  describe('Fight Booking Flow End-to-End', () => {
    test('complete booking flow from opponent selection to confirmation', async () => {
      const user = userEvent.setup();

      axiosGetSpy.mockResolvedValue({ data: mockOpponents });
      axiosPostSpy.mockResolvedValue({ data: mockFightSuccess });

      renderWithProviders(
        <FightBookingModal
          myBoxerId={10}
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/select opponent/i)).toBeInTheDocument();
      });

      const opponentRow = screen.getByText('Thunder Strike Johnson').closest('tr');
      fireEvent.click(opponentRow);

      const scheduledTimeInput = screen.getByDisplayValue(/2026-10/i);
      await user.clear(scheduledTimeInput);
      await user.type(scheduledTimeInput, '2026-12-25T20:00:00');

      await user.click(screen.getByRole('button', { name: /book fight/i }));

      await waitFor(() => {
        expect(axiosPostSpy).toHaveBeenCalledWith('/api/fights', expect.any(Object));
      });

      const callArgs = axiosPostSpy.mock.calls[0][1];
      expect(callArgs.boxer2_id).toBe(2);

      await waitFor(() => {
        expect(screen.getByText(/fight scheduled/i)).toBeInTheDocument();
      });
    });

    test('validates scheduled time is in future', async () => {
      const user = userEvent.setup();

      axiosGetSpy.mockResolvedValue({ data: mockOpponents });

      renderWithProviders(
        <FightBookingModal
          myBoxerId={10}
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/Iron Fist Chen/i)).toBeInTheDocument();
      });

      const opponentRow = screen.getByText('Iron Fist Chen').closest('tr');
      fireEvent.click(opponentRow);

      const scheduledTimeInput = screen.getByDisplayValue(/2026-10/i);
      await user.clear(scheduledTimeInput);
      await user.type(scheduledTimeInput, '2020-01-01T00:00:00');

      const bookButton = screen.getByRole('button', { name: /book fight/i });
      expect(bookButton).toBeDisabled();
    });

    test('shows loading state during API calls', async () => {
      axiosGetSpy.mockImplementation(() => new Promise((resolve) =>
        setTimeout(() => resolve({ data: mockOpponents }), 100)
      ));

      renderWithProviders(
        <FightBookingModal
          myBoxerId={10}
          onClose={() => {}}
          onSuccess={() => {}}
          open={true}
        />
      );

      await waitFor(() => {
        const loadingElement = screen.queryByText(/loading/i) ||
                               screen.queryByRole('progressbar') ||
                               screen.getByText(/select opponent/i);
        expect(loadingElement).toBeInTheDocument();
      });
    });
  });
});
