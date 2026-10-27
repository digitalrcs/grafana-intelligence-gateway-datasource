import React, { ComponentProps, useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { DEFAULT_CONFIG, type Provider } from '../types';
import { ConfigEditor } from './ConfigEditor';

type Options = ComponentProps<typeof ConfigEditor>['options'];

function Harness({
  provider,
  onChange = () => undefined,
}: {
  provider: Provider;
  onChange?: (options: Options) => void;
}) {
  const [options, setOptions] = useState<Options>({
    id: 1,
    uid: 'secure-ai-test',
    name: 'Test secure AI',
    type: 'digitalrcs-intelligencegateway-datasource',
    jsonData: { ...DEFAULT_CONFIG, provider },
    secureJsonData: {},
    secureJsonFields: {},
  } as Options);
  return (
    <ConfigEditor
      options={options}
      onOptionsChange={(next) => {
        setOptions(next);
        onChange(next);
      }}
    />
  );
}

describe('ConfigEditor provider controls', () => {
  test('disables the rendered output-token input for Copilot Studio and shows only its channel credential', () => {
    render(<Harness provider="copilot-studio" />);

    expect(screen.getByRole('spinbutton', { name: /Maximum output tokens/ })).toBeDisabled();
    expect(screen.getByLabelText(/Default agent alias/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Allowed agent aliases/)).toBeInTheDocument();
    expect(screen.getByLabelText('Direct Line secret')).toBeEnabled();
    expect(screen.queryByLabelText('API key')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('Bearer token')).not.toBeInTheDocument();
  });

  test('allows output-token changes for LM Studio and shows its supported credentials', () => {
    const onChange = jest.fn();
    render(<Harness provider="lmstudio" onChange={onChange} />);

    const input = screen.getByRole('spinbutton', { name: /Maximum output tokens/ });
    expect(input).toBeEnabled();
    fireEvent.change(input, { target: { value: '128000' } });
    expect(onChange).toHaveBeenLastCalledWith(
      expect.objectContaining({ jsonData: expect.objectContaining({ maxOutputTokens: 128000 }) })
    );
    expect(screen.getByLabelText('API key')).toBeEnabled();
    expect(screen.getByLabelText('Bearer token')).toBeEnabled();
    expect(screen.queryByLabelText('Direct Line secret')).not.toBeInTheDocument();
  });

  test('updates the disabled state when the provider changes in the rendered form', () => {
    render(<Harness provider="lmstudio" />);
    fireEvent.click(screen.getByRole('radio', { name: 'Copilot Studio' }));
    expect(screen.getByRole('spinbutton', { name: /Maximum output tokens/ })).toBeDisabled();
    expect(screen.getByLabelText('Direct Line secret')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('radio', { name: 'LM Studio' }));
    expect(screen.getByRole('spinbutton', { name: /Maximum output tokens/ })).toBeEnabled();
    expect(screen.getByLabelText('API key')).toBeInTheDocument();
    expect(screen.queryByLabelText('Direct Line secret')).not.toBeInTheDocument();
  });

  test('writes the Direct Line credential only to secureJsonData', () => {
    const onChange = jest.fn();
    render(<Harness provider="copilot-studio" onChange={onChange} />);
    fireEvent.change(screen.getByLabelText('Direct Line secret'), { target: { value: 'test-channel-secret' } });

    const result = onChange.mock.calls[0][0] as Options;
    expect(result.secureJsonData).toEqual({ directLineSecret: 'test-channel-secret' });
    expect(result.jsonData).not.toHaveProperty('directLineSecret');
    expect(result.jsonData.maxOutputTokens).toBe(DEFAULT_CONFIG.maxOutputTokens);
  });
});
