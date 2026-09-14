import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

const API = import.meta.env.VITE_API_URL || "http://localhost:8080";
const CYCLE_MS = 5 * 60 * 60 * 1000;

const durationOf = (item) => Math.max(1, Number(item.duration || 10));

function sequencePosition(playlist, nowMs) {
  if (!playlist.length) return null;

  const cyclePosition = ((nowMs % CYCLE_MS) + CYCLE_MS) % CYCLE_MS;
  const total = playlist.reduce((sum, item) => sum + durationOf(item) * 1000, 0);
  const position = cyclePosition % total;

  let cursor = 0;
  for (let index = 0; index < playlist.length; index++) {
    const duration = durationOf(playlist[index]) * 1000;
    if (position < cursor + duration) {
      return { item: playlist[index], index, remainingMs: cursor + duration - position };
    }
    cursor += duration;
  }
  return { item: playlist[playlist.length - 1], index: playlist.length - 1, remainingMs: 0 };
}

function Display({ windowData, serverNow, sync }) {
  const [position, setPosition] = useState(null);

  useEffect(() => {
    const tick = () => setPosition(sequencePosition(windowData.playlist, serverNow()));
    tick();
    const id = setInterval(tick, 250);
    return () => clearInterval(id);
  }, [windowData.playlist, serverNow]);

  const syncActive = sync?.active && serverNow() < sync.startedAt + sync.duration * 1000;
  const item = syncActive ? sync.media : position?.item;

  return (
    <article className={`display ${syncActive ? "synced" : ""}`}>
      <div className="display-head">
        <strong>{windowData.name}</strong>
        {syncActive && <span>SYNC</span>}
      </div>
      <div className="stage">
        {!item && <div className="empty">No media configured</div>}
        {item?.type === "blank" && <div className="blank">BLANK</div>}
        {item?.type === "image" && <img src={item.url} alt={item.name}/>}
        {item?.type === "video" && <video src={item.url} autoPlay muted loop playsInline/>}
      </div>
      <div className="display-foot">
        <span>{item ? `${item.id} · ${item.name}` : "Waiting"}</span>
        <span>{item ? `${syncActive ? Math.max(0, Math.ceil((sync.startedAt + sync.duration * 1000 - serverNow()) / 1000)) : Math.ceil(position?.remainingMs / 1000)}s` : ""}</span>
      </div>
    </article>
  );
}

function App() {
  const [windows, setWindows] = useState([]);
  const [status, setStatus] = useState("Connecting");
  const [offset, setOffset] = useState(0);
  const [sync, setSync] = useState(null);
  const [selected, setSelected] = useState("");
  const [syncDuration, setSyncDuration] = useState(10);
  const [form, setForm] = useState({ windowId: "window-1", id: "M7", name: "New Media", type: "image", url: "", duration: 10 });

  const serverNow = useMemo(() => () => Date.now() + offset, [offset]);

  useEffect(() => {
    let alive = true;

    async function bootstrap() {
      const before = Date.now();
      const health = await fetch(`${API}/api/health`);
      const after = Date.now();
      if (!health.ok) throw new Error("Backend unavailable");
      const data = await health.json();
      if (alive) {
        // Estimate server clock offset so different browser machines use a common time base.
        setOffset(new Date(data.time || Date.now()).getTime() - Math.round((before + after) / 2));
        setStatus("Live");
      }
      const response = await fetch(`${API}/api/windows`);
      setWindows(await response.json());
    }

    bootstrap().catch(() => alive && setStatus("Backend offline"));

    const events = new EventSource(`${API}/api/events`);
    events.onopen = () => alive && setStatus("Live");
    events.onmessage = (message) => {
      const packet = JSON.parse(message.data);
      if (packet.event === "playlist_updated") {
        setWindows((items) => items.map((w) => w.id === packet.data.id ? packet.data : w));
      }
      if (packet.event === "sync_started") setSync(packet.data);
      if (packet.event === "sync_ended") setSync(null);
    };
    events.onerror = () => alive && setStatus("Reconnecting");

    return () => { alive = false; events.close(); };
  }, []);

  async function addMedia(e) {
    e.preventDefault();
    const response = await fetch(`${API}/api/windows/${form.windowId}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ...form, duration: Number(form.duration) })
    });
    const data = await response.json();
    if (!response.ok) return alert(data.error);
    setWindows((items) => items.map((w) => w.id === data.id ? data : w));
  }

  async function deleteMedia(windowId, mediaId) {
    const media = windows.find((w) => w.id === windowId)?.playlist.find((m) => m.id === mediaId);
    if (!media) return;

    if (!window.confirm(`Delete "${media.name}" from this display window?`)) return;

    const response = await fetch(`${API}/api/windows/${windowId}/media/${encodeURIComponent(mediaId)}`, {
      method: "DELETE"
    });
    const data = await response.json();
    if (!response.ok) return alert(data.error || "Unable to delete media");

    setWindows((items) => items.map((w) => w.id === data.id ? data : w));
    if (selected === mediaId) setSelected("");
  }

  async function syncMedia() {
    const media = windows.flatMap((w) => w.playlist).find((m) => m.id === selected);
    if (!media) return;
    const response = await fetch(`${API}/api/sync`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ media, duration: Number(syncDuration) })
    });
    const data = await response.json();
    if (!response.ok) return alert(data.error);
    setSync(data);
  }

  const uniqueMedia = [...new Map(windows.flatMap((w) => w.playlist).map((m) => [m.id, m])).values()];

  return (
    <main>
      <header>
        <div>
          <small>BACKEND INTERN ASSIGNMENT</small>
          <h1>Multi-Window Media Sequencer</h1>
          <p>Scalable layered backend · React display layer · real-time synchronized playback</p>
        </div>
        <div className={`status ${status === "Live" ? "live" : ""}`}>● {status}</div>
      </header>

      <section className="panels">
        <div className="panel">
          <h2>Add media</h2>
          <form onSubmit={addMedia}>
            <select value={form.windowId} onChange={(e) => setForm({...form, windowId:e.target.value})}>
              {windows.map(w => <option key={w.id} value={w.id}>{w.name}</option>)}
            </select>
            <input placeholder="Media ID" value={form.id} onChange={(e) => setForm({...form,id:e.target.value})}/>
            <input placeholder="Name" value={form.name} onChange={(e) => setForm({...form,name:e.target.value})}/>
            <select value={form.type} onChange={(e) => setForm({...form,type:e.target.value})}>
              <option value="image">Image</option><option value="video">Video</option><option value="blank">Blank</option>
            </select>
            {form.type !== "blank" && <input className="wide" placeholder="Public media URL" value={form.url} onChange={(e) => setForm({...form,url:e.target.value})}/>}
            <input type="number" min="1" value={form.duration} onChange={(e) => setForm({...form,duration:e.target.value})}/>
            <button>Add to playlist</button>
          </form>
        </div>

        <div className="panel">
          <h2>Sync playback</h2>
          <div className="sync-form">
            <select value={selected} onChange={(e) => setSelected(e.target.value)}>
              <option value="">Select media</option>
              {uniqueMedia.map(m => <option key={m.id} value={m.id}>{m.id} · {m.name}</option>)}
            </select>
            <input type="number" min="1" value={syncDuration} onChange={(e) => setSyncDuration(e.target.value)}/>
            <button disabled={!selected} onClick={syncMedia}>Sync all windows</button>
          </div>
          {sync?.active && <div className="note">Syncing <b>{sync.media.name}</b> for {sync.duration}s. Original playlists are unchanged.</div>}
        </div>
      </section>

      <section className="displays">
        {windows.map(w => <Display key={w.id} windowData={w} serverNow={serverNow} sync={sync}/>)}
      </section>

      <section className="lists">
        <h2>Configured playlists</h2>
        <div className="list-grid">
          {windows.map(w => (
            <div className="list" key={w.id}>
              <div className="list-head"><b>{w.name}</b><span>{w.playlist.length} items</span></div>
              {w.playlist.map((m, i) => (
                <div className="media" key={`${m.id}-${i}`}>
                  <button className="media-info" onClick={() => setSelected(m.id)}>
                    <span>{i+1}</span><b>{m.id} · {m.name}</b><small>{m.type} · {durationOf(m)} sec</small>
                  </button>
                  <button
                    className="delete-media"
                    title={`Delete ${m.name}`}
                    onClick={() => deleteMedia(w.id, m.id)}
                  >
                    Delete
                  </button>
                </div>
              ))}
            </div>
          ))}
        </div>
      </section>

      <footer>React · Go · Repository Pattern · Service Layer · SOLID · SSE · Persistent JSON</footer>
    </main>
  );
}

createRoot(document.getElementById("root")).render(<App />);
