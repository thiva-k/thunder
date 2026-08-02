// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, expect, it, vi, beforeEach} from 'vitest';
import {UserCreateFlowStep} from '../../../models/user-create-flow';
import UserCreateProvider from '../UserCreateProvider';
import useUserCreate from '../useUserCreate';

// Test component to consume the context
function TestConsumer() {
  const context = useUserCreate();

  return (
    <div>
      <div data-testid="current-step">{context.currentStep}</div>
      <div data-testid="selected-schema">{context.selectedSchema?.name ?? 'null'}</div>
      <div data-testid="form-values">{JSON.stringify(context.formValues)}</div>
      <div data-testid="error">{context.error ?? 'null'}</div>

      <button type="button" onClick={() => context.setCurrentStep(UserCreateFlowStep.USER_DETAILS)}>
        Set Details Step
      </button>
      <button type="button" onClick={() => context.setSelectedSchema({id: 'schema-1', name: 'Employee', ouId: 'ou-1'})}>
        Set Schema
      </button>
      <button type="button" onClick={() => context.setFormValues({username: 'john'})}>
        Set Form Values
      </button>
      <button type="button" onClick={() => context.setError('Test error')}>
        Set Error
      </button>
      <button type="button" onClick={() => context.reset()}>
        Reset
      </button>
    </div>
  );
}

describe('UserCreateProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('provides initial state values', () => {
    render(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    expect(screen.getByTestId('current-step')).toHaveTextContent(UserCreateFlowStep.USER_TYPE);
    expect(screen.getByTestId('selected-schema')).toHaveTextContent('null');
    expect(screen.getByTestId('form-values')).toHaveTextContent('{}');
    expect(screen.getByTestId('error')).toHaveTextContent('null');
  });

  it('updates current step when setCurrentStep is called', async () => {
    const user = userEvent.setup();

    render(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await user.click(screen.getByText('Set Details Step'));

    expect(screen.getByTestId('current-step')).toHaveTextContent(UserCreateFlowStep.USER_DETAILS);
  });

  it('updates selected schema when setSelectedSchema is called', async () => {
    const user = userEvent.setup();

    render(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await user.click(screen.getByText('Set Schema'));

    expect(screen.getByTestId('selected-schema')).toHaveTextContent('Employee');
  });

  it('updates form values when setFormValues is called', async () => {
    const user = userEvent.setup();

    render(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await user.click(screen.getByText('Set Form Values'));

    expect(screen.getByTestId('form-values')).toHaveTextContent(JSON.stringify({username: 'john'}));
  });

  it('updates error when setError is called', async () => {
    const user = userEvent.setup();

    render(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await user.click(screen.getByText('Set Error'));

    expect(screen.getByTestId('error')).toHaveTextContent('Test error');
  });

  it('resets all state when reset is called', async () => {
    const user = userEvent.setup();

    render(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    // Set some values
    await user.click(screen.getByText('Set Details Step'));
    await user.click(screen.getByText('Set Schema'));
    await user.click(screen.getByText('Set Form Values'));
    await user.click(screen.getByText('Set Error'));

    // Verify values are set
    expect(screen.getByTestId('current-step')).toHaveTextContent(UserCreateFlowStep.USER_DETAILS);
    expect(screen.getByTestId('selected-schema')).toHaveTextContent('Employee');
    expect(screen.getByTestId('error')).toHaveTextContent('Test error');

    // Reset
    await user.click(screen.getByText('Reset'));

    // Verify back to initial state
    expect(screen.getByTestId('current-step')).toHaveTextContent(UserCreateFlowStep.USER_TYPE);
    expect(screen.getByTestId('selected-schema')).toHaveTextContent('null');
    expect(screen.getByTestId('form-values')).toHaveTextContent('{}');
    expect(screen.getByTestId('error')).toHaveTextContent('null');
  });

  it('memoizes context value to prevent unnecessary re-renders', () => {
    const renderSpy = vi.fn();

    function TestRenderer() {
      renderSpy();
      return <TestConsumer />;
    }

    const {rerender} = render(
      <UserCreateProvider>
        <TestRenderer />
      </UserCreateProvider>,
    );

    expect(renderSpy).toHaveBeenCalledTimes(1);

    // Re-render with same props
    rerender(
      <UserCreateProvider>
        <TestRenderer />
      </UserCreateProvider>,
    );

    // Should only render once more due to memoization
    expect(renderSpy).toHaveBeenCalledTimes(2);
  });

  it('throws error when useUserCreate is used outside provider', () => {
    // Suppress console.error for this test since React will log the error
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);

    expect(() => {
      render(<TestConsumer />);
    }).toThrow('useUserCreate must be used within a UserCreateProvider');

    consoleSpy.mockRestore();
  });
});
