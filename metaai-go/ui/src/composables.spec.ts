// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest';
import { useAuth, useLocale } from './composables';



describe('useAuth', () => {
  beforeEach(() => {
    sessionStorage.clear();
    const { logout } = useAuth();
    logout();
  });

  it('starts unauthenticated', () => {
    const { isAuthenticated, token } = useAuth();
    expect(isAuthenticated.value).toBe(false);
    expect(token.value).toBe('');
  });

  it('updates state on login', () => {
    const { isAuthenticated, token, login } = useAuth();
    login('test-token');
    expect(isAuthenticated.value).toBe(true);
    expect(token.value).toBe('test-token');
    expect(sessionStorage.getItem('metaai_token')).toBe('test-token');
  });

  it('clears state on logout', () => {
    const { isAuthenticated, token, login, logout } = useAuth();
    login('test-token');
    logout();
    expect(isAuthenticated.value).toBe(false);
    expect(token.value).toBe('');
    expect(sessionStorage.getItem('metaai_token')).toBe(null);
  });
});

describe('useLocale', () => {
  it('loads default locale', () => {
    const { locale, t } = useLocale();
    expect(locale.value).toBe('en');
    expect(t.value.chat).toBe('Chat');
  });

  it('toggles locale', () => {
    const { locale, t, toggleLocale } = useLocale();
    
    // Toggle to AR
    toggleLocale();
    expect(locale.value).toBe('ar');
    expect(t.value.chat).toBe('محادثة');
    
    // Toggle back to EN
    toggleLocale();
    expect(locale.value).toBe('en');
    expect(t.value.chat).toBe('Chat');
  });
});
