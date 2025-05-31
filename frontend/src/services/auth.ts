import axios from 'axios';

const API_BASE_URL = '/api/v1';

interface LoginResponse {
  user: {
    id: string;
    username: string;
    email: string;
    role: 'player' | 'dm' | 'admin';
  };
  token: string;
}

export const login = async (username: string, password: string): Promise<LoginResponse> => {
  const response = await axios.post(`${API_BASE_URL}/auth/login`, {
    username,
    password,
  });
  return response.data;
};

export const register = async (
  username: string,
  email: string,
  password: string
): Promise<LoginResponse> => {
  const response = await axios.post(`${API_BASE_URL}/auth/register`, {
    username,
    email,
    password,
  });
  return response.data;
};

export const logout = async (): Promise<void> => {
  const token = localStorage.getItem('token');
  if (token) {
    await axios.post(
      `${API_BASE_URL}/auth/logout`,
      {},
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );
  }
};

export const refreshToken = async (): Promise<{ token: string }> => {
  const token = localStorage.getItem('token');
  const response = await axios.post(
    `${API_BASE_URL}/auth/refresh`,
    {},
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  return response.data;
};