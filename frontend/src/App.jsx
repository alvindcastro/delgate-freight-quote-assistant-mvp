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
    try {
      const data = await createQuote(form)
      setQuote(data)
      await refreshQuotes()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleParse() {
    setParsing(true)
    setError('')
    try {
      const data = await parseRequestText(rawText)
      setParseResult(data)
      setForm((current) => mergeDraftIntoForm(current, data.draft, data.extractedFields))
    } catch (err) {
      setError(err.message)
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
      setTimeout(() => setCopied(false), 1600)
    } catch {
      setCopied(false)
      setError('Unable to copy automatically. Select the customer-ready response and copy it manually.')
    }
  }

  return (
    <div className="app-shell">
      <header className="hero">
        <div>
          <p className="eyebrow">DelGate practical assessment</p>
          <h1>AI-powered freight quote assistant</h1>
          <p className="hero-text">
            A Go + React MVP that calculates a transparent demo freight estimate and uses AI to summarize, flag missing details, and draft a customer-ready response.
          </p>
        </div>
        <div className="hero-card">
          <span className="hero-card-label">Design principle</span>
          <strong>Deterministic pricing + AI workflow support</strong>
          <p>The LLM helps with explanation and triage. It does not invent the quote.</p>
        </div>
      </header>

      {error && <div className="alert error">{error}</div>}

      <section className="parser-card">
        <div>
          <p className="section-kicker">Messy request parser</p>
          <h2>Paste a customer freight request</h2>
          <p className="muted">Useful for quote requests that arrive through email, chat, or notes.</p>
        </div>
        <textarea
          value={rawText}
          onChange={(event) => setRawText(event.target.value)}
          rows={3}
          aria-label="Raw customer request"
        />
        <div className="button-row">
          <button className="secondary" type="button" onClick={handleParse} disabled={parsing}>
            {parsing ? 'Parsing…' : 'Parse request'}
          </button>
          <button className="ghost" type="button" onClick={() => setRawText(messyDemo)}>
            Reset sample text
          </button>
        </div>
        {parseResult && (
          <div className="parse-result">
            <div>
              <strong>Extracted:</strong> {parseResult.extractedFields.length ? parseResult.extractedFields.join(', ') : 'No fields detected'}
            </div>
            <div>
              <strong>Still needed:</strong> {parseResult.missingHints.length ? parseResult.missingHints.join(', ') : 'None'}
            </div>
            {parseResult.warnings.length > 0 && <div><strong>Warnings:</strong> {parseResult.warnings.join('; ')}</div>}
          </div>
        )}
      </section>

      <main className="content-grid">
        <form className="quote-form" onSubmit={handleSubmit}>
          <div className="form-title-row">
            <div>
              <p className="section-kicker">Quote intake</p>
              <h2>Shipment details</h2>
            </div>
            <div className="button-row compact">
              <button type="button" className="ghost" onClick={() => setForm(emptyForm)}>Clear</button>
              <button type="button" className="secondary" onClick={() => setForm(demoForm)}>Load demo</button>
            </div>
          </div>

          <div className="form-grid two">
            <Input label="Customer name" value={form.customerName} onChange={(value) => updateField('customerName', value)} />
            <Input label="Customer email" value={form.customerEmail} onChange={(value) => updateField('customerEmail', value)} />
          </div>

          <div className="form-grid two panel-pair">
            <LocationPanel title="Pickup" location={form.origin} onChange={(name, value) => updateLocation('origin', name, value)} />
            <LocationPanel title="Delivery" location={form.destination} onChange={(name, value) => updateLocation('destination', name, value)} />
          </div>

          <div className="form-grid three">
            <Select label="Shipment type" value={form.shipmentType} onChange={(value) => updateField('shipmentType', value)} options={[
              ['ltl_pallet', 'LTL pallet'],
              ['parcel', 'Parcel'],
              ['furniture', 'Furniture'],
              ['appliance', 'Appliance'],
              ['bulky_item', 'Big / bulky item'],
            ]} />
            <Input label="Pieces" type="number" value={form.pieces} onChange={(value) => updateNumber('pieces', value)} />
            <Input label="Pallets" type="number" value={form.pallets} onChange={(value) => updateNumber('pallets', value)} />
          </div>

          <div className="form-grid four">
            <Input label="Weight lbs" type="number" value={form.weightLbs} onChange={(value) => updateNumber('weightLbs', value)} />
            <Input label="Length in" type="number" value={form.lengthIn} onChange={(value) => updateNumber('lengthIn', value)} />
            <Input label="Width in" type="number" value={form.widthIn} onChange={(value) => updateNumber('widthIn', value)} />
            <Input label="Height in" type="number" value={form.heightIn} onChange={(value) => updateNumber('heightIn', value)} />
          </div>

          <Select label="Service level" value={form.serviceLevel} onChange={(value) => updateField('serviceLevel', value)} options={[
            ['standard', 'Standard'],
            ['expedited', 'Expedited'],
            ['same_day', 'Same day'],
          ]} />

          <fieldset className="accessorials">
            <legend>Accessorials</legend>
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

          <label className="field full">
            <span>Operations notes</span>
            <textarea
              value={form.notes}
              onChange={(event) => updateField('notes', event.target.value)}
              rows={4}
              placeholder="Example: Customer needs delivery before Friday; confirm appointment window."
            />
          </label>

          <button className="primary" type="submit" disabled={loading}>
            {loading ? 'Generating quote…' : 'Generate quote'}
          </button>
        </form>

        <aside className="result-column">
          {quote ? (
            <QuoteResult quote={quote} currency={currency} onCopy={copyCustomerMessage} copied={copied} />
          ) : (
            <div className="empty-state">
              <p className="section-kicker">Quote result</p>
              <h2>No quote yet</h2>
              <p>Load the demo or enter shipment details to generate the first estimate.</p>
            </div>
          )}

          <QuoteHistory quotes={quotes} currency={currency} onSelect={setQuote} />
        </aside>
      </main>
    </div>
  )
}

function Input({ label, value, onChange, type = 'text' }) {
  return (
    <label className="field">
      <span>{label}</span>
      <input type={type} value={value} onChange={(event) => onChange(event.target.value)} />
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
  return (
    <div className="location-panel">
      <h3>{title}</h3>
      <Input label="City" value={location.city} onChange={(value) => onChange('city', value)} />
      <Input label="Province" value={location.province} onChange={(value) => onChange('province', value)} />
      <Input label="Postal code" value={location.postalCode} onChange={(value) => onChange('postalCode', value)} />
    </div>
  )
}

function QuoteResult({ quote, currency, onCopy, copied }) {
  return (
    <section className="quote-result">
      <div className="quote-topline">
        <div>
          <p className="section-kicker">{quote.quoteId}</p>
          <h2>{quote.routeLabel}</h2>
        </div>
        <StatusBadge value={quote.status} />
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

      <section className="mini-section">
        <h3>Breakdown</h3>
        <BreakdownTable breakdown={quote.breakdown} currency={currency} />
      </section>

      <section className="mini-section">
        <h3>AI operations summary</h3>
        <p className="summary-box">{quote.assistantSummary}</p>
      </section>

      <section className="mini-section">
        <h3>Missing details</h3>
        <TagList items={quote.missingFields} empty="No missing details detected" />
      </section>

      <section className="mini-section">
        <h3>Risk flags</h3>
        <TagList items={quote.riskFlags} empty="No risk flags detected" />
      </section>

      <section className="mini-section">
        <div className="section-title-row">
          <h3>Customer-ready response</h3>
          <button type="button" className="secondary small" onClick={onCopy}>{copied ? 'Copied' : 'Copy'}</button>
        </div>
        <pre className="message-box">{quote.customerMessage}</pre>
      </section>
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
    <section className="history-card">
      <div className="section-title-row">
        <div>
          <p className="section-kicker">In-memory history</p>
          <h2>Recent quotes</h2>
        </div>
      </div>
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
    </section>
  )
}

function StatusBadge({ value }) {
  const className = value === 'Ready to Quote' ? 'ready' : value === 'Manual Review Required' ? 'manual' : 'review'
  return <span className={`status-badge ${className}`}>{value}</span>
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
