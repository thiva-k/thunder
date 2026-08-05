// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Components
export {default as ExternalLinkConfirmDialog} from './ExternalLinkConfirm/ExternalLinkConfirmDialog';
export type {ExternalLinkConfirmDialogProps} from './ExternalLinkConfirm/ExternalLinkConfirmDialog';
export {default as useExternalLinkConfirmation} from './ExternalLinkConfirm/useExternalLinkConfirmation';
export type {ExternalLinkConfirmationState} from './ExternalLinkConfirm/useExternalLinkConfirmation';
export {default as Helmet} from './Helmet/Helmet';
export type {HelmetProps} from './Helmet/Helmet';
export {default as FullScreenCreationWizardLayout} from './FullScreenCreationWizardLayout/FullScreenCreationWizardLayout';
export type {FullScreenCreationWizardLayoutProps} from './FullScreenCreationWizardLayout/FullScreenCreationWizardLayout';
export {default as I18nTextInput} from './I18nTextInput/I18nTextInput';
export type {I18nTextInputLabels, I18nTextInputProps} from './I18nTextInput/I18nTextInput';
export {default as ExternalLink} from './ExternalLink/ExternalLink';
export type {ExternalLinkProps} from './ExternalLink/ExternalLink';
export {default as OrganizationUnitSummaryChip} from './OrganizationUnitSummaryChip/OrganizationUnitSummaryChip';
export type {OrganizationUnitSummaryChipProps} from './OrganizationUnitSummaryChip/OrganizationUnitSummaryChip';
export {default as PageLoader} from './PageLoader/PageLoader';
export {default as ReadErrorState} from './ReadErrorState/ReadErrorState';
export type {ReadErrorStateProps, ReadErrorMessageResolver} from './ReadErrorState/ReadErrorState';
export {default as ToggleCard} from './ToggleCard/ToggleCard';
export type {ToggleCardProps} from './ToggleCard/ToggleCard';

/* -------------------------- ICONS -------------------------- */

export {default as GithubIcon} from './icons/logos/vendor/GithubIcon';
export {default as GoogleIcon} from './icons/logos/vendor/GoogleIcon';
export {default as HeidiIcon} from './icons/logos/vendor/HeidiIcon';
export {default as LissiIcon} from './icons/logos/vendor/LissiIcon';

/* -------------------------- LAB -------------------------- */

export {default as BuilderFloatingPanel} from './lab/components/BuilderLayout/BuilderFloatingPanel';
export {default as BuilderLayout} from './lab/components/BuilderLayout/BuilderLayout';
export {default as BuilderPanelHeader} from './lab/components/BuilderLayout/BuilderPanelHeader';
export {default as BuilderStaticPanel} from './lab/components/BuilderLayout/BuilderStaticPanel';
export {default as EmojiPicker} from './lab/components/EmojiPicker/EmojiPicker';
export {default as CopyableId} from './lab/components/CopyableId';
export {default as Kbd} from './lab/components/Kbd';
export {default as generateIconSuggestions} from './lab/components/EmojiPicker/utils/generateIconSuggestions';
export {default as LogoPicker} from './lab/components/LogoPicker/LogoPicker';
export type {LogoPickerProps} from './lab/components/LogoPicker/LogoPicker';
export {default as NameSuggestion} from './lab/components/NameSuggestion';
export type {NameSuggestionProps} from './lab/components/NameSuggestion';
export {default as PageLoadingAnimation} from './lab/components/PageLoadingAnimation';
export {default as ResourceAvatar} from './lab/components/ResourceAvatar';
export {default as SettingsCard} from './lab/components/SettingsCard';
export {default as UnsavedChangesBar} from './lab/components/UnsavedChangesBar';

// Utils
export {default as getInitials} from './lab/utils/getInitials';
