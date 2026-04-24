import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../api';
import { useAuth } from '../AuthContext';
import { useWebSocket } from '../useWebSocket';
import type { MedicalCase, Phase, Classroom } from '../types';

export default function ClassroomPage() {
  const { code } = useParams<{ code: string }>();

  // If no code, show classroom list/create
  if (!code) return <ClassroomList />;
  return <ClassroomViewer code={code} />;
}

function ClassroomList() {
  const { user } = useAuth();
  const nav = useNavigate();
  const [classrooms, setClassrooms] = useState<Classroom[]>([]);
  const [joinCode, setJoinCode] = useState('');
  const [cases, setCases] = useState<MedicalCase[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState({ case_id: 0, title: '' });

  useEffect(() => {
    api.listClassrooms().then(setClassrooms).catch(() => {});
    api.getCases().then(setCases).catch(() => {});
  }, []);

  const handleJoin = () => { if (joinCode.trim()) nav(`/classrooms/${joinCode.trim()}`); };
  const handleCreate = async () => {
    if (!createForm.case_id || !createForm.title) return;
    const res = await api.createClassroom(createForm.case_id, createForm.title);
    nav(`/classrooms/${res.code}`);
  };

  return (
    <div className="container" style={{ paddingTop: 40, paddingBottom: 60 }}>
      <h1 className="page-title">📡 Virtual Classrooms</h1>
      <p className="page-subtitle">Real-time synchronized medical image viewing sessions</p>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24, marginBottom: 40 }}>
        {/* Join */}
        <div className="card">
          <h3 style={{ marginBottom: 16 }}>Join a Classroom</h3>
          <div style={{ display: 'flex', gap: 12 }}>
            <input className="input" placeholder="Enter 6-digit code" value={joinCode}
              onChange={e => setJoinCode(e.target.value.toUpperCase())} maxLength={6}
              style={{ fontFamily: 'monospace', fontSize: 18, letterSpacing: 3, textAlign: 'center' }} />
            <button className="btn btn-primary" onClick={handleJoin}>Join</button>
          </div>
        </div>

        {/* Create (doctor only) */}
        {user?.role === 'doctor' && (
          <div className="card">
            <h3 style={{ marginBottom: 16 }}>Create Classroom</h3>
            {!showCreate ? (
              <button className="btn btn-primary" onClick={() => setShowCreate(true)}>+ New Session</button>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                <input className="input" placeholder="Session title" value={createForm.title}
                  onChange={e => setCreateForm({...createForm, title: e.target.value})} />
                <select className="input" value={createForm.case_id}
                  onChange={e => setCreateForm({...createForm, case_id: Number(e.target.value)})}>
                  <option value={0}>Select a case...</option>
                  {cases.map(c => <option key={c.id} value={c.id}>{c.title}</option>)}
                </select>
                <button className="btn btn-primary" onClick={handleCreate}>Start Session</button>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Active classrooms */}
      <h2 style={{ fontSize: 20, marginBottom: 16 }}>Active Sessions</h2>
      {classrooms.length === 0 ? (
        <div style={{ color: 'var(--text-muted)', padding: 40, textAlign: 'center' }}>No active sessions</div>
      ) : (
        <div className="cases-grid">
          {classrooms.map(cr => (
            <div key={cr.id} className="card case-card" onClick={() => nav(`/classrooms/${cr.code}`)}>
              <div className="case-card-header">
                <span className="live-badge"><span className="live-dot"></span> LIVE</span>
                <span className="classroom-code" style={{ fontSize: 14, padding: '2px 8px' }}>{cr.code}</span>
              </div>
              <div className="case-card-body">
                <h3>{cr.title}</h3>
                <div style={{ color: 'var(--text-muted)', fontSize: 13, marginTop: 8 }}>
                  🩺 {cr.doctor?.name} • {cr.case?.title || 'Medical Case'}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function ClassroomViewer({ code }: { code: string }) {
  const { user } = useAuth();
  const nav = useNavigate();
  const [classroom, setClassroom] = useState<Classroom | null>(null);
  const [mc, setMc] = useState<MedicalCase | null>(null);
  const [activePhase, setActivePhase] = useState<Phase | null>(null);
  const [sliceIndex, setSliceIndex] = useState(0);
  const [onlineUsers, setOnlineUsers] = useState<string[]>([]);
  const [cursorPos, setCursorPos] = useState<{x: number; y: number; name: string} | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [dataLoaded, setDataLoaded] = useState(false);
  const imgRef = useRef<HTMLDivElement>(null);
  const isDoctor = user?.role === 'doctor';

  // Refs to avoid stale closures in WS handlers
  const mcRef = useRef<MedicalCase | null>(null);
  const isDoctorRef = useRef(isDoctor);
  mcRef.current = mc;
  isDoctorRef.current = isDoctor;

  // Only connect WebSocket AFTER data has loaded successfully
  const wsUrl = dataLoaded ? api.wsUrl(code) : null;
  const { connected, send, on } = useWebSocket(wsUrl);

  // Load classroom data
  useEffect(() => {
    console.log('[Classroom] Loading data for code:', code);
    api.joinClassroom(code).then(data => {
      console.log('[Classroom] Data loaded:', data);
      setClassroom(data.classroom);
      const caseData = data.classroom?.case;
      if (caseData) {
        setMc(caseData);
        mcRef.current = caseData;
        if (caseData.phases?.length > 0) {
          setActivePhase(caseData.phases[0]);
        }
      }
      setOnlineUsers(
        Array.isArray(data.online_users)
          ? data.online_users.map((u: any) => typeof u === 'string' ? u : u.name || '')
          : []
      );
      setDataLoaded(true);
    }).catch((err) => {
      console.error('[Classroom] Failed to load:', err);
      setError(err.message || 'Failed to load classroom');
    });
  }, [code]);

  // Listen for WebSocket messages — use refs to avoid stale closures
  useEffect(() => {
    if (!dataLoaded) return;

    const unsub1 = on('slice:changed', (msg: any) => {
      if (!isDoctorRef.current) setSliceIndex(msg.payload?.slice_index ?? 0);
    });
    const unsub2 = on('phase:changed', (msg: any) => {
      if (!isDoctorRef.current && mcRef.current) {
        const phase = mcRef.current.phases?.find((p: Phase) => p.id === msg.payload?.phase_id);
        if (phase) {
          setActivePhase(phase);
          setSliceIndex(0);
        }
      }
    });
    const unsub3 = on('cursor:moved', (msg: any) => {
      if (!isDoctorRef.current) setCursorPos({ x: msg.payload?.x, y: msg.payload?.y, name: msg.user_name || 'Doctor' });
    });
    // Hub sends full user list on join/leave
    const unsub4 = on('presence:update', (msg: any) => {
      const users = msg.payload as Array<{id: number; name: string}>;
      if (Array.isArray(users)) {
        setOnlineUsers(users.map(u => u.name));
      }
    });
    const unsub5 = on('user:joined', (msg: any) => {
      setOnlineUsers(prev => [...new Set([...prev, msg.user_name || ''])]);
    });
    const unsub6 = on('user:left', (msg: any) => {
      setOnlineUsers(prev => prev.filter(n => n !== msg.user_name));
    });
    return () => { unsub1(); unsub2(); unsub3(); unsub4(); unsub5(); unsub6(); };
  }, [on, dataLoaded]);

  // Doctor: broadcast slice changes
  const handleSliceChange = useCallback((newIndex: number) => {
    setSliceIndex(newIndex);
    if (isDoctor) {
      send({ type: 'slice:changed', payload: { slice_index: newIndex, phase_id: activePhase?.id } });
    }
  }, [isDoctor, send, activePhase]);

  // Doctor: broadcast phase changes
  const handlePhaseChange = useCallback((phase: Phase) => {
    setActivePhase(phase); setSliceIndex(0);
    if (isDoctor) {
      send({ type: 'phase:changed', payload: { phase_id: phase.id } });
      send({ type: 'slice:changed', payload: { slice_index: 0, phase_id: phase.id } });
    }
  }, [isDoctor, send]);

  // Doctor: broadcast cursor position
  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!isDoctor || !imgRef.current) return;
    const rect = imgRef.current.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width;
    const y = (e.clientY - rect.top) / rect.height;
    send({ type: 'cursor:moved', payload: { x, y, phase_id: activePhase?.id } });
  }, [isDoctor, send, activePhase]);

  // Wheel scroll
  useEffect(() => {
    const el = imgRef.current;
    if (!el) return;
    const handleWheel = (e: WheelEvent) => {
      e.preventDefault();
      if (!activePhase?.slices) return;
      const next = sliceIndex + (e.deltaY > 0 ? 1 : -1);
      handleSliceChange(Math.max(0, Math.min(activePhase.slices.length - 1, next)));
    };
    el.addEventListener('wheel', handleWheel, { passive: false });
    return () => el.removeEventListener('wheel', handleWheel);
  }, [sliceIndex, activePhase, handleSliceChange]);

  // Error state
  if (error) return (
    <div style={{ padding: 60, textAlign: 'center' }}>
      <div style={{ color: 'var(--accent-red)', fontSize: 18, marginBottom: 16 }}>⚠️ {error}</div>
      <button className="btn btn-secondary" onClick={() => nav('/classrooms')}>← Back to Classrooms</button>
    </div>
  );

  // Loading state
  if (!mc || !classroom) return (
    <div style={{ padding: 60, textAlign: 'center', color: 'var(--text-muted)' }}>
      <div style={{ fontSize: 18, marginBottom: 8 }}>Loading classroom...</div>
      <div style={{ fontSize: 13 }}>Connecting to {code}</div>
    </div>
  );

  const currentSlice = activePhase?.slices?.[sliceIndex];
  const imageUrl = currentSlice ? api.mediaUrl(mc.folder_name, activePhase!.folder_name, currentSlice.file_name) : '';
  const totalSlices = activePhase?.slices?.length || 0;

  return (
    <div className="viewer-page">
      <div className="viewer-sidebar">
        <button className="btn btn-secondary btn-sm" onClick={() => nav('/classrooms')} style={{ marginBottom: 16 }}>← Back</button>

        <div className="live-badge" style={{ marginBottom: 16 }}><span className="live-dot"></span> LIVE SESSION</div>
        <h3 style={{ marginBottom: 8 }}>{classroom.title}</h3>
        <div className="classroom-code" style={{ marginBottom: 16 }}>{classroom.code}</div>

        <div style={{ fontSize: 13, color: connected ? 'var(--accent-emerald)' : 'var(--accent-red)', marginBottom: 16 }}>
          {connected ? '● Connected' : '○ Connecting...'}
        </div>

        {isDoctor && (
          <div style={{ padding: 8, background: 'rgba(59,130,246,0.1)', borderRadius: 8, fontSize: 12, color: 'var(--accent)', marginBottom: 16 }}>
            👨‍⚕️ You are the Doctor — students follow your view
          </div>
        )}
        {!isDoctor && (
          <div style={{ padding: 8, background: 'rgba(6,182,212,0.1)', borderRadius: 8, fontSize: 12, color: 'var(--accent-cyan)', marginBottom: 16 }}>
            🎓 Following doctor's view in real-time
          </div>
        )}

        <h4 style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 8 }}>Online ({onlineUsers.length})</h4>
        <div className="online-users">
          {onlineUsers.map((name, i) => (
            <div key={i} className="online-user"><span className="online-dot"></span> {name}</div>
          ))}
        </div>

        <div className="viewer-info" style={{ marginTop: 20 }}>
          <h3 style={{ fontSize: 15 }}>{mc.title}</h3>
          <div style={{ margin: '8px 0' }}>
            <span className={`modality-badge ${mc.modality.toLowerCase()}`}>{mc.modality}</span>
            <span style={{ marginLeft: 8, fontSize: 13, color: 'var(--text-muted)' }}>{mc.body_part}</span>
          </div>
          <div className="diagnosis-text" style={{ marginTop: 12 }}>{mc.diagnosis}</div>
        </div>
      </div>

      <div className="viewer-main">
        <div className="viewer-toolbar">
          <div className="phase-tabs">
            {mc.phases?.map(p => (
              <button key={p.id} className={`phase-tab ${activePhase?.id === p.id ? 'active' : ''}`}
                onClick={() => handlePhaseChange(p)} disabled={!isDoctor}>
                {p.name}
              </button>
            ))}
          </div>
          <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
            {isDoctor ? '🩺 Doctor Mode — Scroll to control' : '🎓 Student Mode — Synced'}
          </div>
        </div>

        <div className="viewer-image-container" ref={imgRef} onMouseMove={handleMouseMove}>
          {imageUrl && <img src={imageUrl} alt={`Slice ${sliceIndex + 1}`} draggable={false} />}
          <div className="slice-counter">Slice {sliceIndex + 1} / {totalSlices}</div>

          {/* Remote cursor overlay */}
          {!isDoctor && cursorPos && (
            <div className="remote-cursor" data-name={cursorPos.name}
              style={{ left: `${cursorPos.x * 100}%`, top: `${cursorPos.y * 100}%` }} />
          )}
        </div>

        <div className="slice-slider-container">
          <input type="range" className="slice-slider" min={0} max={Math.max(0, totalSlices - 1)}
            value={sliceIndex}
            onChange={e => handleSliceChange(Number(e.target.value))}
            disabled={!isDoctor} />
        </div>
      </div>
    </div>
  );
}
