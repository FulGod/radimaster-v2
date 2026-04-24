import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../api';
import type { MedicalCase, Phase } from '../types';

export default function ViewerPage() {
  const { id } = useParams<{ id: string }>();
  const nav = useNavigate();
  const [mc, setMc] = useState<MedicalCase | null>(null);
  const [sliceIndex, setSliceIndex] = useState(0);
  const [brightness, setBrightness] = useState(1);
  const [contrast, setContrast] = useState(1);
  const viewerRef = useRef<HTMLDivElement>(null);

  // Split-view: select 2 phases to display side by side
  const [leftPhase, setLeftPhase] = useState<Phase | null>(null);
  const [rightPhase, setRightPhase] = useState<Phase | null>(null);

  useEffect(() => {
    if (!id) return;
    api.getCase(Number(id)).then(data => {
      setMc(data);
      const phases = data.phases || [];
      // Auto-select first 2 phases for split-view
      if (phases.length >= 2) {
        setLeftPhase(phases[0]);
        setRightPhase(phases[1]);
      } else if (phases.length === 1) {
        setLeftPhase(phases[0]);
        setRightPhase(null);
      }
      setSliceIndex(0);
    }).catch(() => nav('/'));
  }, [id]);

  // Preload images for both phases
  useEffect(() => {
    if (!mc) return;
    [leftPhase, rightPhase].forEach(phase => {
      if (!phase) return;
      phase.slices?.forEach(s => {
        const img = new Image();
        img.src = api.mediaUrl(mc.folder_name, phase.folder_name, s.file_name);
      });
    });
  }, [leftPhase, rightPhase, mc]);

  // Max slices across both panes
  const maxSlices = Math.max(
    leftPhase?.slices?.length || 0,
    rightPhase?.slices?.length || 0
  );

  // Mouse wheel slice scrolling — synced across both panes
  const handleWheel = useCallback((e: WheelEvent) => {
    e.preventDefault();
    setSliceIndex(prev => {
      const next = prev + (e.deltaY > 0 ? 1 : -1);
      return Math.max(0, Math.min(maxSlices - 1, next));
    });
  }, [maxSlices]);

  useEffect(() => {
    const el = viewerRef.current;
    if (!el) return;
    el.addEventListener('wheel', handleWheel, { passive: false });
    return () => el.removeEventListener('wheel', handleWheel);
  }, [handleWheel]);

  // Keyboard navigation
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'ArrowDown' || e.key === 'ArrowRight') {
        e.preventDefault();
        setSliceIndex(p => Math.min(maxSlices - 1, p + 1));
      } else if (e.key === 'ArrowUp' || e.key === 'ArrowLeft') {
        e.preventDefault();
        setSliceIndex(p => Math.max(0, p - 1));
      }
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [maxSlices]);

  if (!mc) return <div style={{ padding: 60, textAlign: 'center', color: 'var(--text-muted)' }}>Loading...</div>;

  const phases = mc.phases || [];
  const hasSplitView = rightPhase !== null;

  return (
    <div className="viewer-page">
      {/* Sidebar */}
      <div className="viewer-sidebar">
        <button className="btn btn-secondary btn-sm" onClick={() => nav('/')} style={{ marginBottom: 20 }}>
          ← Back to Library
        </button>
        <div className="viewer-info">
          <h2>{mc.title}</h2>
          <div style={{ margin: '12px 0' }}>
            <span className={`modality-badge ${mc.modality.toLowerCase()}`}>{mc.modality}</span>
            <span style={{ marginLeft: 8, fontSize: 13, color: 'var(--text-muted)' }}>{mc.body_part}</span>
          </div>
          <div className="meta-item"><span className="meta-label">Doctor</span><span>🩺 {mc.doctor?.name}</span></div>
          <div className="meta-item"><span className="meta-label">Phases</span><span>{phases.length}</span></div>
          <div className="meta-item"><span className="meta-label">Views</span><span>{mc.view_count}</span></div>

          <h3 style={{ marginTop: 20, marginBottom: 8, fontSize: 14, color: 'var(--accent-cyan)' }}>Diagnosis</h3>
          <div className="diagnosis-text">{mc.diagnosis}</div>

          <h3 style={{ marginTop: 16, marginBottom: 8, fontSize: 14, color: 'var(--text-secondary)' }}>Description</h3>
          <div className="desc-text">{mc.description}</div>

          {/* Phase selection for split-view */}
          {phases.length >= 2 && (
            <div style={{ marginTop: 24 }}>
              <div style={{ fontSize: 12, color: 'var(--text-muted)', marginBottom: 8 }}>Split-View Phases</div>
              <div style={{ display: 'flex', gap: 8, marginBottom: 8, alignItems: 'center' }}>
                <span style={{ fontSize: 11, color: 'var(--accent-cyan)', minWidth: 40 }}>Left</span>
                <select className="input" style={{ fontSize: 12, padding: '4px 8px' }}
                  value={leftPhase?.id || ''}
                  onChange={e => { const p = phases.find(ph => ph.id === Number(e.target.value)); if (p) { setLeftPhase(p); setSliceIndex(0); } }}>
                  {phases.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
              <div style={{ display: 'flex', gap: 8, marginBottom: 8, alignItems: 'center' }}>
                <span style={{ fontSize: 11, color: 'var(--accent-emerald)', minWidth: 40 }}>Right</span>
                <select className="input" style={{ fontSize: 12, padding: '4px 8px' }}
                  value={rightPhase?.id || ''}
                  onChange={e => { const p = phases.find(ph => ph.id === Number(e.target.value)); if (p) { setRightPhase(p); setSliceIndex(0); } }}>
                  {phases.map(p => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
            </div>
          )}

          {/* Brightness/Contrast Controls */}
          <div style={{ marginTop: 24 }}>
            <div style={{ fontSize: 12, color: 'var(--text-muted)', marginBottom: 8 }}>Image Controls</div>
            <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 8 }}>
              <span style={{ fontSize: 12, color: 'var(--text-secondary)', minWidth: 70 }}>Brightness</span>
              <input type="range" min="0.2" max="3" step="0.1" value={brightness}
                onChange={e => setBrightness(Number(e.target.value))} className="slice-slider" />
            </div>
            <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
              <span style={{ fontSize: 12, color: 'var(--text-secondary)', minWidth: 70 }}>Contrast</span>
              <input type="range" min="0.2" max="3" step="0.1" value={contrast}
                onChange={e => setContrast(Number(e.target.value))} className="slice-slider" />
            </div>
            <button className="btn btn-secondary btn-sm" style={{ marginTop: 8 }} onClick={() => { setBrightness(1); setContrast(1); }}>
              Reset
            </button>
          </div>
        </div>
      </div>

      {/* Main Viewer — Split View */}
      <div className="viewer-main" ref={viewerRef}>
        {/* Phase labels */}
        <div className="viewer-toolbar">
          <div className="phase-tabs">
            {hasSplitView ? (
              <>
                <span style={{ fontSize: 13, color: 'var(--accent-cyan)' }}>◀ {leftPhase?.name}</span>
                <span style={{ fontSize: 12, color: 'var(--text-muted)', margin: '0 8px' }}>|</span>
                <span style={{ fontSize: 13, color: 'var(--accent-emerald)' }}>{rightPhase?.name} ▶</span>
              </>
            ) : (
              <span style={{ fontSize: 13, color: 'var(--accent-cyan)' }}>{leftPhase?.name}</span>
            )}
          </div>
          <div style={{ fontSize: 13, color: 'var(--text-muted)' }}>
            Scroll ↕ to navigate — both panes sync
          </div>
        </div>

        {/* Split Image Display */}
        <div className={`viewer-split-container ${hasSplitView ? 'dual' : 'single'}`}
          style={{ '--brightness': brightness, '--contrast': contrast } as React.CSSProperties}>

          {/* Left pane */}
          <div className="viewer-pane">
            <div className="pane-label left">{leftPhase?.name}</div>
            {leftPhase && mc && (() => {
              const slice = leftPhase.slices?.[sliceIndex];
              const url = slice ? api.mediaUrl(mc.folder_name, leftPhase.folder_name, slice.file_name) : '';
              return url ? <img src={url} alt={`Left ${sliceIndex + 1}`} draggable={false} /> : null;
            })()}
            <div className="slice-counter">
              {sliceIndex + 1} / {leftPhase?.slices?.length || 0}
            </div>
          </div>

          {/* Right pane (if split-view) */}
          {hasSplitView && rightPhase && (
            <div className="viewer-pane">
              <div className="pane-label right">{rightPhase.name}</div>
              {mc && (() => {
                const slice = rightPhase.slices?.[sliceIndex];
                const url = slice ? api.mediaUrl(mc.folder_name, rightPhase.folder_name, slice.file_name) : '';
                return url ? <img src={url} alt={`Right ${sliceIndex + 1}`} draggable={false} /> : null;
              })()}
              <div className="slice-counter">
                {sliceIndex + 1} / {rightPhase.slices?.length || 0}
              </div>
            </div>
          )}
        </div>

        {/* Slice Slider (synced) */}
        <div className="slice-slider-container">
          <input type="range" className="slice-slider" min={0} max={Math.max(0, maxSlices - 1)}
            value={sliceIndex} onChange={e => setSliceIndex(Number(e.target.value))} />
        </div>
      </div>
    </div>
  );
}
