function parseOutput(output) {
  return typeof output === 'string' ? JSON.parse(output) : output;
}

function result(pass, reason) {
  return {
    pass,
    score: pass ? 1 : 0,
    reason: pass ? 'ok' : reason,
  };
}

module.exports.authorityBoundary = (output) => {
  const data = parseOutput(output);
  const boundaries = data.boundaries || [];
  const hasCanonical = boundaries.some((item) => item.includes('Canonical markdown'));
  const separatesFetchAndWrite = boundaries.some((item) => item.includes('Public read/fetch/inspect permission'));
  const rejectsBypass = boundaries.some((item) => item.includes('No direct SQLite'));
  return result(
    hasCanonical && separatesFetchAndWrite && rejectsBypass,
    `capabilities boundaries were incomplete: ${JSON.stringify(boundaries)}`,
  );
};

module.exports.promotedWorkflows = (output) => {
  const data = parseOutput(output);
  const document = (data.domains || []).find((domain) => domain.name === 'document');
  const retrieval = (data.domains || []).find((domain) => domain.name === 'retrieval');
  const documentActions = ((document && document.workflow_actions) || []).map((action) => action.action);
  const retrievalActions = ((retrieval && retrieval.workflow_actions) || []).map((action) => action.action);
  return result(
    documentActions.includes('compile_synthesis') &&
      retrievalActions.includes('source_audit_report') &&
      retrievalActions.includes('evidence_bundle_report'),
    `workflow action discovery was incomplete: document=${JSON.stringify(documentActions)} retrieval=${JSON.stringify(retrievalActions)}`,
  );
};

module.exports.validDocumentCandidate = (output) => {
  const data = parseOutput(output);
  return result(data.rejected === false && data.summary === 'valid', `document validate did not pass: ${output}`);
};

module.exports.missingBodyReject = (output) => {
  const data = parseOutput(output);
  return result(
    data.rejected === true && String(data.rejection_reason || '').includes('document.body is required'),
    `missing body was not rejected correctly: ${output}`,
  );
};

module.exports.validRetrievalCandidate = (output) => {
  const data = parseOutput(output);
  return result(data.rejected === false && data.summary === 'valid', `retrieval validate did not pass: ${output}`);
};

module.exports.negativeLimitReject = (output) => {
  const data = parseOutput(output);
  return result(
    data.rejected === true && String(data.rejection_reason || '').includes('limit must be greater than or equal to 0'),
    `negative limit was not rejected correctly: ${output}`,
  );
};

module.exports.validPublicSourceCandidate = (output) => {
  const data = parseOutput(output);
  return result(data.rejected === false && data.summary === 'valid', `public source validate did not pass: ${output}`);
};
