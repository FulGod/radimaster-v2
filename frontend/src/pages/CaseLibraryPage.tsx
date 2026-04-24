import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api';
import type { MedicalCase } from '../types';

export default function CaseLibraryPage() {
  const [cases, setCases] = useState<MedicalCase[]>([]);
  const [modalities, setModalities] = useState<string[]>([]);
  const [activeModality, setActiveModality] = useState('');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const nav = useNavigate();

  useEffect(() => { api.getModalities().then(setModalities).catch(() => {}); }, []);

  useEffect(() => {
    setLoading(true);
    api.getCases(activeModality || undefined, search || undefined)
      .then(setCases).finally(() => setLoading(false));
  }, [activeModality, search]);

  const modalityColors: Record<string, string> = { CT: 'ct', MRI: 'mri', XRay: 'xray', Ultrasound: 'mri' };

  return (
    <div className="container" style={{ paddingTop: 40, paddingBottom: 60 }}>
      <h1 className="page-title">📋 Case Library</h1>
      <p className="page-subtitle">Browse real medical imaging cases from hospital practice</p>

      <div style={{ marginBottom: 24 }}>
        <input className="input" placeholder="🔍 Search by title, diagnosis, or body part..." value={search} onChange={e => setSearch(e.target.value)} style={{ maxWidth: 480 }} />
      </div>

      <div className="filters">
        <button className={`filter-btn ${!activeModality ? 'active' : ''}`} onClick={() => setActiveModality('')}>All</button>
        {modalities.map(m => (
          <button key={m} className={`filter-btn ${activeModality === m ? 'active' : ''}`} onClick={() => setActiveModality(m)}>{m}</button>
        ))}
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 60, color: 'var(--text-muted)' }}>Loading cases...</div>
      ) : cases.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 60, color: 'var(--text-muted)' }}>No cases found</div>
      ) : (
        <div className="cases-grid">
          {cases.map(c => (
            <div key={c.id} className="card case-card" onClick={() => nav(`/cases/${c.id}`)}>
              <div className="case-card-header">
                <span className={`modality-badge ${modalityColors[c.modality] || 'ct'}`}>{c.modality}</span>
                <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>{c.body_part}</span>
              </div>
              <div className="case-card-body">
                <h3>{c.title}</h3>
                <div className="diagnosis">{c.diagnosis}</div>
                <div className="case-card-meta">
                  <span>🩺 {c.doctor?.name || 'Doctor'}</span>
                  <span>📂 {c.phases?.length || 0} phases</span>
                  <span>👁 {c.view_count} views</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
