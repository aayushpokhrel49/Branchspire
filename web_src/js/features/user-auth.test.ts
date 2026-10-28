import {initUserAuthPasswordToggle} from './user-auth.ts';

test('initUserAuthPasswordToggle', () => {
  document.body.innerHTML = `
    <div class="auth-password-field">
      <input id="password" type="password" value="secret">
      <button type="button" class="auth-password-toggle" data-target="password" data-show-label="Show password" data-hide-label="Hide password" aria-label="Show password">
        <span class="auth-icon-show">show</span><span class="auth-icon-hide">hide</span>
      </button>
    </div>`;
  initUserAuthPasswordToggle();

  const btn = document.querySelector('.auth-password-toggle');
  btn.click();
  const input = document.querySelector('#password');
  expect(input.type).toEqual('text');
  expect(btn.classList.contains('is-visible')).toEqual(true);
  expect(btn.getAttribute('aria-label')).toEqual('Hide password');

  btn.click();
  expect(input.type).toEqual('password');
  expect(btn.classList.contains('is-visible')).toEqual(false);
  expect(btn.getAttribute('aria-label')).toEqual('Show password');
});
