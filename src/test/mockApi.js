/**
 * Mock API Layer for Testing
 *
 * Provides mock implementations of API functions for isolated component testing.
 * Configure mock responses before tests using jest.spyOn or direct assignments.
 *
 * @see MAT-169 - Frontend Component Tests
 */

// Default mock data
const defaultBoxers = [
  {
    id: 1,
    name: 'Mike Tyson',
    nickname: 'Iron Mike',
    strength: 85,
    defense: 70,
    agility: 90,
    health: 100,
    energy: 100,
    fatigue_score: 20,
    experience: 5000,
    level: 50,
    wins: 50,
    losses: 5,
    draws: 0,
    has_active_rest: false,
  },
  {
    id: 2,
    name: 'Evander Holyfield',
    nickname: 'The Real Deal',
    strength: 90,
    defense: 85,
    agility: 75,
    health: 100,
    energy: 100,
    fatigue_score: 15,
    experience: 6000,
    level: 55,
    wins: 45,
    losses: 3,
    draws: 2,
    has_active_rest: false,
  },
];

const defaultOpponents = [
  {
    boxer_id: 2,
    boxer_name: 'Evander Holyfield',
    strength: 90,
    defense: 85,
    agility: 75,
    match_score: 85,
    win_probability: 0.45,
  },
];

const defaultTrainingTypes = [
  { id: 1, name: 'strength_training', label: 'Strength', icon: '💪' },
  { id: 2, name: 'speed_training', label: 'Speed', icon: '⚡' },
  { id: 3, name: 'defense_training', label: 'Defense', icon: '🛡️' },
  { id: 4, name: 'technique_training', label: 'Technique', icon: '🥊' },
  { id: 5, name: 'endurance_training', label: 'Endurance', icon: '🏃' },
];

// Mock API functions - can be overridden in tests
export const mockApi = {
  // Boxer endpoints
  getBoxers: jest.fn().mockResolvedValue(defaultBoxers),
  getBoxer: jest.fn().mockImplementation((id) => {
    return Promise.resolve(defaultBoxers.find(b => b.id === id) || null);
  }),

  // Fight endpoints
  getUpcomingFights: jest.fn().mockResolvedValue([]),
  getFightHistory: jest.fn().mockResolvedValue([]),
  bookFight: jest.fn().mockResolvedValue({ success: true }),

  // Training endpoints
  getTrainingTypes: jest.fn().mockResolvedValue(defaultTrainingTypes),
  getActiveTrainingSessions: jest.fn().mockResolvedValue([]),
  scheduleTraining: jest.fn().mockResolvedValue({ success: true }),

  // Opponent discovery
  discoverOpponents: jest.fn().mockResolvedValue(defaultOpponents),

  // Rankings
  getRankings: jest.fn().mockResolvedValue([]),

  // User auth
  login: jest.fn().mockResolvedValue({ token: 'mock-jwt-token' }),
  register: jest.fn().mockResolvedValue({ success: true }),
};

/**
 * Reset all mock functions to default state
 * Call this in beforeEach() for clean test isolation
 */
export function resetMocks() {
  Object.values(mockApi).forEach(mock => {
    if (typeof mock === 'function' && mock.mock) {
      mock.mockReset();
    }
  });
}

/**
 * Setup module mocks for API imports
 * Use this in your test setup to automatically mock API calls
 */
export function setupModuleMocks() {
  // Mock the entire api module
  jest.mock('../api', () => ({
    ...jest.requireActual('../api'),
    fetchBoxers: mockApi.getBoxers,
    fetchBoxer: mockApi.getBoxer,
    fetchUpcomingFight: mockApi.getUpcomingFights,
    fetchActiveTrainingSessions: mockApi.getActiveTrainingSessions,
  }));
}

export default mockApi;
