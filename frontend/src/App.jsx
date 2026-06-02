import { useEffect, useMemo, useState } from 'react'
import { createQuote, getQuotes, parseRequestText } from './api.js'

const accessorialOptions = [
  { value: 'liftgate', label: 'Liftgate' },
  { value: 'residential', label: 'Residential delivery' },
  { value: 'inside_delivery', label: 'Inside delivery' },
  { value: 'appointment_required', label: 'Appointment required' },
  { value: 'limited_access', label: 'Limited access' },
  { value: 'fragile', label: 'Fragile' },
  { value: 'oversized', label: 'Oversized' },
  { value: 'weekend', label: 'Weekend' },
]

const emptyForm = {
  customerName: '',
  customerEmail: '',
  origin: { city: '', province: '', postalCode: '' },
  destination: { city: '', province: '', postalCode: '' },
  shipmentType: 'ltl_pallet',
  pieces: 1,
  pallets: 1,
  weightLbs: 500,
  lengthIn: 48,
  widthIn: 40,
  heightIn: 60,
  serviceLevel: 'standard',
  accessorials: [],
  notes: '',
}

const demoForm = {
  customerName: 'Daniel',
  customerEmail: 'ops@example.com',
  origin: { city: 'Vancouver', province: 'BC', postalCode: 'V6B 1A1' },
  destination: { city: 'Calgary', province: 'AB', postalCode: 'T2P 1J9' },
  shipmentType: 'ltl_pallet',
  pieces: 2,
  pallets: 2,
  weightLbs: 700,
  lengthIn: 48,
  widthIn: 40,
  heightIn: 60,
  serviceLevel: 'standard',
  accessorials: ['liftgate', 'residential'],
  notes: 'Customer prefers delivery before Friday. Confirm appointment window and whether inside delivery is needed.',
}

const messyDemo = 'Need to ship 2 pallets from Vancouver to Calgary. Each pallet is 48x40x60 and 350 lbs. Customer needs liftgate and residential delivery before Friday.'

function App() {
  const [form, setForm] = useState(emptyForm)
  const [quote, setQuote] = useState(null)
  const [quotes, setQuotes] = useState([])
  const [rawText, setRawText] = useState(messyDemo)
  const [parseResult, setParseResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [parsing, setParsing] = useState(false)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState(false)
  const [liveMessage, setLiveMessage] = useState('')
  const [parserOpen, setParserOpen] = useState(false)

  useEffect(() => {
    refreshQuotes()
  }, [])

  const currency = useMemo(
    () => new Intl.NumberFormat('en-CA', { style: 'currency', currency: 'CAD', maximumFractionDigits: 0 }),
    [],
  )

  async function refreshQuotes() {
    try {
      const data = await getQuotes()
      setQuotes(data)
    } catch {
      // Ignore history errors during initial local startup.
    }
  }

  function updateField(name, value) {
    setForm((current) => ({ ...current, [name]: value }))
  }

  function updateLocation(type, name, value) {
    setForm((current) => ({
      ...current,
      [type]: {
        ...current[type],
        [name]: value,
      },
    }))
  }

  function updateNumber(name, value) {
    const parsed = Number(value)
    updateField(name, Number.isNaN(parsed) ? 0 : parsed)
  }

  function updateRawText(value) {
    setRawText(value)
    if (parseResult) {
      setParseResult(null)
    }
  }

  function resetParserText() {
    setRawText(messyDemo)
    setParseResult(null)
    setParserOpen(true)
  }

  function clearParserText() {
    setRawText('')
    setParseResult(null)
    setError('')
    setLiveMessage('Parser text cleared.')
  }

  function clearForm() {
    setForm(emptyForm)
    setError('')
    setLiveMessage('Shipment form cleared.')
  }

  function clearQuoteResult() {
    setQuote(null)
    setCopied(false)
    setError('')
    setLiveMessage('Quote result pane cleared.')
  }

  function clearWorkspace() {
    setForm(emptyForm)
    setQuote(null)
    setRawText('')
    setParseResult(null)
    setError('')
    setCopied(false)
    setLiveMessage('Workspace cleared.')
  }

  function toggleAccessorial(value) {
    setForm((current) => {
      const exists = current.accessorials.includes(value)
      return {
        ...current,
        accessorials: exists
          ? current.accessorials.filter((item) => item !== value)
          : [...current.accessorials, value],
      }
    })
  }

  async function handleSubmit(event) {
    event.preventDefault()
    setLoading(true)
    setError('')
    setCopied(false)
    setLiveMessage('Generating quote.')
    try {
      const data = await createQuote(form)
      setQuote(data)
      await refreshQuotes()
      setLiveMessage('Quote generated.')
    } catch (err) {
      setError(err.message)
      setLiveMessage('Quote generation failed.')
    } finally {
      setLoading(false)
    }
  }

  async function handleParse() {
    setParsing(true)
    setError('')
    setLiveMessage('Parsing customer request.')
    try {
      const data = await parseRequestText(rawText)
      setParseResult(data)
      setForm((current) => mergeDraftIntoForm(current, data.draft, data.extractedFields))
      setLiveMessage('Customer request parsed.')
    } catch (err) {
      setError(err.message)
      setLiveMessage('Request parsing failed.')
    } finally {
      setParsing(false)
    }
  }

  async function copyCustomerMessage() {
    if (!quote?.customerMessage) return
    setError('')
    try {
      if (!navigator.clipboard?.writeText) {
        throw new Error('Clipboard API unavailable')
      }
      await navigator.clipboard.writeText(quote.customerMessage)
      setCopied(true)
      setLiveMessage('Customer response copied.')
      setTimeout(() => setCopied(false), 1600)
    } catch {
      setCopied(false)
      setError('Unable to copy automatically. Select the customer-ready response and copy it manually.')
      setLiveMessage('Copy failed.')
    }
  }

  return (
    <div className="app-shell">
      <header className="hero">
        <div className="hero-copy">
          <SectionLabel text="DelGate workbench" />
          <h1>
            Freight quotes with <span className="gradient-text">operator control</span>
          </h1>
          <p className="hero-text">
            Estimate, review, and draft customer-ready freight responses from one focused workspace.
          </p>
          <div className="hero-actions">
            <button type="button" className="primary action-fit" onClick={() => setForm(demoForm)}>
              Load demo
              <span aria-hidden="true">→</span>
            </button>
            <button type="button" className="secondary action-fit" onClick={clearWorkspace}>
              Clear workspace
            </button>
          </div>
        </div>
        <HeroGraphic quote={quote} quotes={quotes} />
      </header>

      {error && (
        <div className="alert error" role="alert" aria-live="polite">
          {error}
        </div>
      )}

      <div className="sr-only" role="status" aria-live="polite">{liveMessage}</div>

      <details
        className="parser-card disclosure-card"
        open={parserOpen}
        onToggle={(event) => setParserOpen(event.currentTarget.open)}
        aria-busy={parsing}
      >
        <summary className="disclosure-summary">
          <div>
            <SectionLabel text="Request parser" />
            <h2 id="parser-title">Customer text parser</h2>
            <p>Paste messy shipment details only when you need assisted intake.</p>
          </div>
          <div className="summary-actions">
            {parseResult && <span className="result-count">{parseResult.extractedFields.length} fields</span>}
            <span className="summary-toggle" aria-hidden="true" />
          </div>
        </summary>
        <div className="disclosure-body" aria-labelledby="parser-title">
          <label className="field full parser-input" htmlFor="raw-request-text">
            <span>Request text</span>
            <textarea
              id="raw-request-text"
              value={rawText}
              onChange={(event) => updateRawText(event.target.value)}
              rows={3}
            />
          </label>
          <div className="button-row">
            <button className="primary action-fit" type="button" onClick={handleParse} disabled={parsing || !rawText.trim()}>
              {parsing ? 'Parsing...' : 'Parse request'}
              <span aria-hidden="true">→</span>
            </button>
            <button className="secondary action-fit" type="button" onClick={resetParserText}>
              Demo text
            </button>
            <button className="ghost action-fit" type="button" onClick={clearParserText} disabled={!rawText && !parseResult}>
              Clear text
            </button>
          </div>
          {parseResult && (
            <div className="parse-result" aria-live="polite">
              <ParseGroup label="Extracted" items={parseResult.extractedFields} empty="No fields detected" />
              <ParseGroup label="Still needed" items={parseResult.missingHints} empty="None" />
              {parseResult.warnings.length > 0 && <ParseGroup label="Warnings" items={parseResult.warnings} />}
            </div>
          )}
        </div>
      </details>

      <main className="content-grid">
        <form className="quote-form" onSubmit={handleSubmit} aria-busy={loading}>
          <div className="form-title-row">
            <div>
              <SectionLabel text="Quote intake" />
              <h2>Shipment details</h2>
            </div>
            <div className="button-row compact">
              <button type="button" className="ghost action-fit" onClick={clearForm}>Clear form</button>
              <button type="button" className="secondary action-fit" onClick={() => setForm(demoForm)}>Load demo</button>
            </div>
          </div>

          <DisclosureSection title="Customer" defaultOpen>
            <div className="form-grid two">
              <Input label="Customer name" value={form.customerName} autoComplete="name" onChange={(value) => updateField('customerName', value)} />
              <Input label="Customer email" type="email" value={form.customerEmail} autoComplete="email" onChange={(value) => updateField('customerEmail', value)} />
            </div>
          </DisclosureSection>

          <DisclosureSection title="Route" defaultOpen>
            <div className="form-grid two panel-pair">
              <LocationPanel title="Pickup" location={form.origin} onChange={(name, value) => updateLocation('origin', name, value)} />
              <LocationPanel title="Delivery" location={form.destination} onChange={(name, value) => updateLocation('destination', name, value)} />
            </div>
          </DisclosureSection>

          <DisclosureSection title="Freight profile" defaultOpen>
            <div className="form-grid three">
              <Select label="Shipment type" value={form.shipmentType} onChange={(value) => updateField('shipmentType', value)} options={[
                ['ltl_pallet', 'LTL pallet'],
                ['parcel', 'Parcel'],
                ['furniture', 'Furniture'],
                ['appliance', 'Appliance'],
                ['bulky_item', 'Big / bulky item'],
              ]} />
              <Input label="Pieces" type="number" value={form.pieces} min="0" step="1" inputMode="numeric" onChange={(value) => updateNumber('pieces', value)} />
              <Input label="Pallets" type="number" value={form.pallets} min="0" step="1" inputMode="numeric" onChange={(value) => updateNumber('pallets', value)} />
            </div>

            <div className="form-grid four">
              <Input label="Weight lbs" type="number" value={form.weightLbs} min="0" step="any" inputMode="decimal" onChange={(value) => updateNumber('weightLbs', value)} />
              <Input label="Length in" type="number" value={form.lengthIn} min="0" step="any" inputMode="decimal" onChange={(value) => updateNumber('lengthIn', value)} />
              <Input label="Width in" type="number" value={form.widthIn} min="0" step="any" inputMode="decimal" onChange={(value) => updateNumber('widthIn', value)} />
              <Input label="Height in" type="number" value={form.heightIn} min="0" step="any" inputMode="decimal" onChange={(value) => updateNumber('heightIn', value)} />
            </div>

            <Select label="Service level" value={form.serviceLevel} onChange={(value) => updateField('serviceLevel', value)} options={[
              ['standard', 'Standard'],
              ['expedited', 'Expedited'],
              ['same_day', 'Same day'],
            ]} />
          </DisclosureSection>

          <DisclosureSection title="Accessorials">
            <fieldset className="accessorials">
              <legend className="sr-only">Accessorials</legend>
              <div className="checkbox-grid">
                {accessorialOptions.map((item) => (
                  <label key={item.value} className="checkbox-card">
                    <input
                      type="checkbox"
                      checked={form.accessorials.includes(item.value)}
                      onChange={() => toggleAccessorial(item.value)}
                    />
                    <span>{item.label}</span>
                  </label>
                ))}
              </div>
            </fieldset>
          </DisclosureSection>

          <DisclosureSection title="Operations notes">
            <label className="field full">
              <span>Notes</span>
              <textarea
                value={form.notes}
                onChange={(event) => updateField('notes', event.target.value)}
                rows={3}
                placeholder="Customer needs delivery before Friday; confirm appointment window."
              />
            </label>
          </DisclosureSection>

          <button className="primary submit-button" type="submit" disabled={loading}>
            {loading ? 'Generating quote…' : 'Generate quote'}
            <span aria-hidden="true">→</span>
          </button>
        </form>

        <aside className="result-column">
          {quote ? (
            <QuoteResult quote={quote} currency={currency} onCopy={copyCustomerMessage} copied={copied} onClear={clearQuoteResult} />
          ) : (
            <div className="empty-state">
              <SectionLabel text="Quote result" />
              <h2>No quote yet</h2>
              <p>Load the demo or enter shipment details to generate the first estimate.</p>
            </div>
          )}

          <QuoteHistory quotes={quotes} currency={currency} onSelect={setQuote} />
        </aside>
      </main>

      <section className="ops-strip" aria-label="Quote workflow status">
        <div>
          <SectionLabel text="Operations rhythm" inverted />
          <h2>Deterministic pricing. AI-assisted communication.</h2>
        </div>
        <div className="ops-metrics">
          <Metric label="Pricing mode" value="Rules first" />
          <Metric label="History cap" value="1,000" />
          <Metric label="AI guardrail" value="Fallback ready" />
        </div>
      </section>
    </div>
  )
}

function SectionLabel({ text, inverted = false }) {
  return (
    <p className={`section-kicker${inverted ? ' inverted' : ''}`}>
      <span aria-hidden="true" />
      {text}
    </p>
  )
}

function HeroGraphic({ quote, quotes }) {
  const latestStatus = quote?.status || 'Ready when details land'
  const summary = quote
    ? `Current quote estimate ${quote.currency} ${Math.round(quote.midpoint)}, status ${latestStatus}, ${quotes.length} recent quotes.`
    : `${quotes.length} recent quotes. No current quote estimate yet.`
  return (
    <div className="hero-graphic" role="img" aria-label={summary}>
      <div className="hero-ring" aria-hidden="true" />
      <div className="hero-node primary-node" aria-hidden="true" />
      <div className="hero-node secondary-node" aria-hidden="true" />
      <div className="hero-panel floating-one">
        <span>Estimate</span>
        <strong>{quote ? `${quote.currency} ${Math.round(quote.midpoint)}` : 'CAD ready'}</strong>
      </div>
      <div className="hero-panel floating-two">
        <span>Status</span>
        <strong>{latestStatus}</strong>
      </div>
      <div className="hero-grid-dots" aria-hidden="true" />
      <div className="hero-count">
        <span>{quotes.length}</span>
        <small>recent</small>
      </div>
    </div>
  )
}

function DisclosureSection({ title, children, defaultOpen = false }) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <details className="form-section disclosure-section" open={open} onToggle={(event) => setOpen(event.currentTarget.open)}>
      <summary className="mini-summary">
        <h3>{title}</h3>
        <span className="summary-toggle" aria-hidden="true" />
      </summary>
      <div className="disclosure-body">{children}</div>
    </details>
  )
}

function ParseGroup({ label, items, empty = '' }) {
  return (
    <div>
      <strong>{label}</strong>
      <div className="parse-chip-list">
        {items.length ? items.map((item) => <span key={item}>{item}</span>) : <em>{empty}</em>}
      </div>
    </div>
  )
}

function Input({ label, value, onChange, type = 'text', ...props }) {
  return (
    <label className="field">
      <span>{label}</span>
      <input type={type} value={value} onChange={(event) => onChange(event.target.value)} {...props} />
    </label>
  )
}

function Select({ label, value, onChange, options }) {
  return (
    <label className="field">
      <span>{label}</span>
      <select value={value} onChange={(event) => onChange(event.target.value)}>
        {options.map(([optionValue, optionLabel]) => (
          <option key={optionValue} value={optionValue}>{optionLabel}</option>
        ))}
      </select>
    </label>
  )
}

function LocationPanel({ title, location, onChange }) {
  const section = title.toLowerCase()
  return (
    <fieldset className="location-panel">
      <legend>{title}</legend>
      <Input label={`${title} city`} value={location.city} autoComplete={`section-${section} address-level2`} onChange={(value) => onChange('city', value)} />
      <Input label={`${title} province`} value={location.province} autoComplete={`section-${section} address-level1`} onChange={(value) => onChange('province', value)} />
      <Input label={`${title} postal code`} value={location.postalCode} autoComplete={`section-${section} postal-code`} onChange={(value) => onChange('postalCode', value)} />
    </fieldset>
  )
}

function QuoteResult({ quote, currency, onCopy, copied, onClear }) {
  return (
    <section className="quote-result">
      <div className="quote-topline">
        <div>
          <SectionLabel text={quote.quoteId} />
          <h2>{quote.routeLabel}</h2>
        </div>
        <div className="quote-actions">
          <StatusBadge value={quote.status} />
          <button type="button" className="ghost small" onClick={onClear}>Clear result</button>
        </div>
      </div>

      <div className="estimate-card">
        <span>Estimated range</span>
        <strong>{currency.format(quote.estimatedLow)} – {currency.format(quote.estimatedHigh)}</strong>
        <p>{quote.disclaimer}</p>
      </div>

      <div className="metric-grid">
        <Metric label="Confidence" value={quote.confidence} />
        <Metric label="Chargeable weight" value={`${Math.round(quote.chargeableWeightLbs)} lbs`} />
        <Metric label="AI provider" value={quote.aiProvider} />
      </div>

      <DisclosureSection title="Breakdown">
        <BreakdownTable breakdown={quote.breakdown} currency={currency} />
      </DisclosureSection>

      <DisclosureSection title="AI operations summary">
        <p className="summary-box">{quote.assistantSummary}</p>
      </DisclosureSection>

      <DisclosureSection title="Missing details">
        <TagList items={quote.missingFields} empty="No missing details detected" />
      </DisclosureSection>

      <DisclosureSection title="Risk flags">
        <TagList items={quote.riskFlags} empty="No risk flags detected" />
      </DisclosureSection>

      <DisclosureSection title="Customer-ready response" defaultOpen>
        <div className="copy-row">
          <button type="button" className="secondary small" onClick={onCopy}>{copied ? 'Copied' : 'Copy response'}</button>
        </div>
        <pre className="message-box">{quote.customerMessage}</pre>
      </DisclosureSection>
    </section>
  )
}

function BreakdownTable({ breakdown, currency }) {
  const rows = [
    ['Lane base', breakdown.laneBase],
    ['Weight charge', breakdown.weightCharge],
    ['Dimensional charge', breakdown.dimensionalCharge],
    ['Piece handling', breakdown.pieceHandling],
    ['Shipment surcharge', breakdown.shipmentSurcharge],
    ['Service adjustment', breakdown.serviceAdjustment],
    ['Accessorial fees', breakdown.accessorialFees],
    ['Fuel surcharge', breakdown.fuelSurcharge],
  ]
  return (
    <table className="breakdown-table">
      <tbody>
        {rows.map(([label, value]) => (
          <tr key={label}>
            <td>{label}</td>
            <td>{currency.format(value || 0)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function QuoteHistory({ quotes, currency, onSelect }) {
  return (
    <details className="history-card disclosure-card">
      <summary className="disclosure-summary">
        <div>
          <SectionLabel text="In-memory history" />
          <h2>Recent quotes</h2>
        </div>
        <div className="summary-actions">
          <span className="result-count">{quotes.length} saved</span>
          <span className="summary-toggle" aria-hidden="true" />
        </div>
      </summary>
      <div className="disclosure-body">
        {quotes.length === 0 ? (
          <p className="muted">Generated quotes will appear here.</p>
        ) : (
          <div className="history-list">
            {quotes.map((item) => (
              <button key={item.quoteId} type="button" onClick={() => onSelect(item)} className="history-item">
                <span>{item.routeLabel}</span>
                <strong>{currency.format(item.estimatedLow)} – {currency.format(item.estimatedHigh)}</strong>
                <small>{item.status}</small>
              </button>
            ))}
          </div>
        )}
      </div>
    </details>
  )
}

function StatusBadge({ value }) {
  const className = value === 'Ready to Quote' ? 'ready' : value === 'Manual Review Required' ? 'manual' : 'review'
  return (
    <span className={`status-badge ${className}`}>
      <span aria-hidden="true" />
      {value}
    </span>
  )
}

function Metric({ label, value }) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

function TagList({ items, empty }) {
  if (!items || items.length === 0) {
    return <p className="muted">{empty}</p>
  }
  return (
    <div className="tag-list">
      {items.map((item) => <span key={item} className="tag">{item}</span>)}
    </div>
  )
}

function mergeDraftIntoForm(current, draft = {}, extractedFields = []) {
  const accessorials = Array.isArray(draft.accessorials) ? draft.accessorials : []
  const extracted = new Set(extractedFields.map((field) => String(field).toLowerCase()))
  const has = (...fields) => fields.some((field) => extracted.has(field))
  const pickString = (value, fallback) => (typeof value === 'string' && value.trim() !== '' ? value : fallback)
  const pickNumber = (value, fallback) => (Number.isFinite(value) && value > 0 ? value : fallback)

  return {
    ...current,
    customerName: has('customer name') ? pickString(draft.customerName, current.customerName) : current.customerName,
    customerEmail: has('customer email') ? pickString(draft.customerEmail, current.customerEmail) : current.customerEmail,
    origin: {
      city: has('origin', 'origin city') ? pickString(draft.origin?.city, current.origin.city) : current.origin.city,
      province: has('origin', 'origin province') ? pickString(draft.origin?.province, current.origin.province) : current.origin.province,
      postalCode: has('origin postal code', 'pickup postal code') ? pickString(draft.origin?.postalCode, current.origin.postalCode) : current.origin.postalCode,
    },
    destination: {
      city: has('destination', 'destination city') ? pickString(draft.destination?.city, current.destination.city) : current.destination.city,
      province: has('destination', 'destination province') ? pickString(draft.destination?.province, current.destination.province) : current.destination.province,
      postalCode: has('destination postal code', 'delivery postal code') ? pickString(draft.destination?.postalCode, current.destination.postalCode) : current.destination.postalCode,
    },
    shipmentType: has('shipment type', 'pallet count') ? pickString(draft.shipmentType, current.shipmentType) : current.shipmentType,
    pieces: has('piece count', 'pallet count') ? pickNumber(draft.pieces, current.pieces) : current.pieces,
    pallets: has('pallet count') ? pickNumber(draft.pallets, current.pallets) : current.pallets,
    weightLbs: has('weight') ? pickNumber(draft.weightLbs, current.weightLbs) : current.weightLbs,
    lengthIn: has('dimensions') ? pickNumber(draft.lengthIn, current.lengthIn) : current.lengthIn,
    widthIn: has('dimensions') ? pickNumber(draft.widthIn, current.widthIn) : current.widthIn,
    heightIn: has('dimensions') ? pickNumber(draft.heightIn, current.heightIn) : current.heightIn,
    serviceLevel: has('service level') ? pickString(draft.serviceLevel, current.serviceLevel) : current.serviceLevel,
    accessorials: has('accessorials') && accessorials.length ? accessorials : current.accessorials,
    notes: current.notes ? current.notes : pickString(draft.notes, current.notes),
  }
}

export default App
