import { useState, type FormEvent } from 'react';
import { fetchCheckRelationship, type Note } from '../../lib/graphql';
import { Button } from '../common/Button';
import { ErrorState } from '../common/ErrorState';

export interface RelationshipCheckerProps {
  notes: Note[];
  isAiUnavailable: boolean;
  onDetectedAiUnavailable: () => void;
}

export function RelationshipChecker({
  notes,
  isAiUnavailable,
  onDetectedAiUnavailable,
}: RelationshipCheckerProps) {
  const [noteIdA, setNoteIdA] = useState('');
  const [noteIdB, setNoteIdB] = useState('');
  const [analysis, setAnalysis] = useState<string | null>(null);
  const [isChecking, setIsChecking] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  if (isAiUnavailable || notes.length < 2) {
    return null;
  }

  const handleCheck = async (event: FormEvent) => {
    event.preventDefault();
    if (!noteIdA || !noteIdB || noteIdA === noteIdB) return;

    setIsChecking(true);
    setErrorMessage(null);
    setAnalysis(null);

    try {
      const response = await fetchCheckRelationship(noteIdA, noteIdB);
      if (response.toLowerCase().includes('ai features are disabled')) {
        onDetectedAiUnavailable();
        return;
      }
      setAnalysis(response);
    } catch (err) {
      const msg =
        err instanceof Error
          ? err.message
          : 'Unable to check relationship between notes.';
      if (msg.toLowerCase().includes('ai client not enabled')) {
        onDetectedAiUnavailable();
        return;
      }
      setErrorMessage(msg);
    } finally {
      setIsChecking(false);
    }
  };

  return (
    <section
      className="relationship-checker-panel"
      aria-labelledby="relationship-checker-heading"
    >
      <div className="relationship-checker-header">
        <h2 id="relationship-checker-heading" className="context-section-heading">
          Compare Note Relationship
        </h2>
        <p className="page-description">
          Select any two notes to inspect their conceptual relationship using local AI.
        </p>
      </div>

      <form className="relationship-form" onSubmit={(e) => void handleCheck(e)}>
        <div className="relationship-select-group">
          <label htmlFor="rel-note-a" className="input-label">
            First Note
          </label>
          <select
            id="rel-note-a"
            className="select-input"
            value={noteIdA}
            onChange={(e) => setNoteIdA(e.target.value)}
            disabled={isChecking}
          >
            <option value="">Select first note…</option>
            {notes.map((n) => (
              <option key={n.id} value={n.id}>
                {n.title || 'Untitled Note'}
              </option>
            ))}
          </select>
        </div>

        <div className="relationship-select-group">
          <label htmlFor="rel-note-b" className="input-label">
            Second Note
          </label>
          <select
            id="rel-note-b"
            className="select-input"
            value={noteIdB}
            onChange={(e) => setNoteIdB(e.target.value)}
            disabled={isChecking}
          >
            <option value="">Select second note…</option>
            {notes
              .filter((n) => n.id !== noteIdA)
              .map((n) => (
                <option key={n.id} value={n.id}>
                  {n.title || 'Untitled Note'}
                </option>
              ))}
          </select>
        </div>

        <div className="relationship-submit-wrap">
          <Button
            type="submit"
            variant="secondary"
            disabled={isChecking || !noteIdA || !noteIdB || noteIdA === noteIdB}
          >
            {isChecking ? 'Analyzing…' : 'Check Relationship'}
          </Button>
        </div>
      </form>

      {errorMessage && (
        <ErrorState
          title="Relationship check failed"
          message={errorMessage}
        />
      )}

      {analysis && (
        <div className="relationship-result-card" role="region" aria-label="Relationship analysis">
          <span className="context-section-heading">AI Relationship Analysis</span>
          <p className="smart-explanation-body">{analysis}</p>
        </div>
      )}
    </section>
  );
}
