import React, { ComponentProps, useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { DEFAULT_CONFIG, type IntelligenceGatewayDataSourceOptions, type Provider } from '../types';
import { ConfigEditor } from './ConfigEditor';

type Options = ComponentProps<typeof ConfigEditor>['options'];

function Harness({
  provider,
  jsonData = {},
  onChange = () => undefined,
}: {
  provider: Provider;
  jsonData?: Partial<IntelligenceGatewayDataSourceOptions>;
  onChange?: (options: Options) => void;
}) {
  const [options, setOptions] = useState<Options>({
    id: 1,
    uid: 'secure-ai-test',
    name: 'Test secure AI',
    type: 'digitalrcs-intelligencegateway-datasource',
    jsonData: { ...DEFAULT_CONFIG, provider, ...jsonData },
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
  const mockTitle = 'DEMO ONLY - Mock AI (no inference)';
  const mockFixture = { reviewMock: true, baseUrl: 'http://mock-ai:8080/v1' };

  test('shows the marked mock connection notice and a separate real-provider checklist', () => {
    render(<Harness provider="custom" jsonData={mockFixture} />);

    expect(screen.getByText(mockTitle)).toBeInTheDocument();
    expect(screen.getByText(/This demo returns a fixed delivery receipt/)).toBeInTheDocument();
    expect(screen.getByText('Create a separate Intelligence Gateway Secure AI data source.')).toBeInTheDocument();
    expect(screen.getByText(/not its Queries selector/)).toBeInTheDocument();
    expect(screen.queryByRole('checkbox', { name: /mock/i })).not.toBeInTheDocument();
  });

  test('normalizes surrounding whitespace and trailing slashes for the marked mock URL', () => {
    render(<Harness provider="custom" jsonData={{ reviewMock: true, baseUrl: ' http://mock-ai:8080/v1/ ' }} />);

    expect(screen.getByText(mockTitle)).toBeInTheDocument();
  });

  test.each([undefined, false])('does not label an unmarked or disabled instance as mock (%s)', (reviewMock) => {
    render(<Harness provider="custom" jsonData={{ baseUrl: mockFixture.baseUrl, reviewMock }} />);

    expect(screen.queryByText(mockTitle)).not.toBeInTheDocument();
    expect(DEFAULT_CONFIG).not.toHaveProperty('reviewMock');
  });

  test('clears the notice when the provider changes even if the fixture marker remains', () => {
    render(<Harness provider="custom" jsonData={mockFixture} />);
    expect(screen.getByText(mockTitle)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('radio', { name: 'LM Studio' }));
    expect(screen.queryByText(mockTitle)).not.toBeInTheDocument();
  });

  test('clears the notice when the base URL changes even if the fixture marker remains', () => {
    render(<Harness provider="custom" jsonData={mockFixture} />);
    expect(screen.getByText(mockTitle)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/Base URL/), {
      target: { value: 'https://ai.example.com/v1' },
    });
    expect(screen.queryByText(mockTitle)).not.toBeInTheDocument();
  });

  test.each(['https://mock-ai:8080/v1', 'http://mock-ai:8080/v1/other', 'http://mock-ai:8080/v1?real=true'])(
    'does not label another marked endpoint as the supplied mock: %s',
    (baseUrl) => {
      render(<Harness provider="custom" jsonData={{ reviewMock: true, baseUrl }} />);
      expect(screen.queryByText(mockTitle)).not.toBeInTheDocument();
    }
  );

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
