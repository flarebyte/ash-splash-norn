type KeyStructKind = string //note:text:comment
type validationPhase = 'BeforeValidation' | 'AfterValidation'

type Validation = {
    args: string[];
    stop: boolean;//if fails
    tags: string[];//warning
    phase: validationPhase;
}

type Formatting = {
    args: string[];//trim
    tags: string[];//warning
}

export type ValidationSchema = {
  validationByKeyStruct: Record<KeyStructKind, Validation[]>;
  formattingByKeyStruct: Record<KeyStructKind, Formatting[]>;
};
