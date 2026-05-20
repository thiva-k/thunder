/**
 * No-op shim for @asgardeo/react.
 *
 * The frontend design package dist imports these symbols from @asgardeo/react,
 * but docs only uses theme utilities from that package. This shim prevents
 * webpack from failing when it cannot resolve @asgardeo/react in the docs build.
 */
module.exports = {
  Consent: () => null,
  ConsentCheckboxList: () => null,
  EmbeddedFlowComponentType: {},
  EmbeddedFlowEventType: {},
  EmbeddedFlowTextVariant: {},
  FlowTimer: () => null,
  extractEmojiFromUri: () => undefined,
  isEmojiUri: () => false,
  useAsgardeo: () => ({}),
};
