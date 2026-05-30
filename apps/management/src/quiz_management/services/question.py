from collections.abc import Sequence
from datetime import UTC, datetime
from uuid import UUID

from sqlalchemy.ext.asyncio.session import AsyncSession

from quiz_management.core.exceptions import ServiceException
from quiz_management.models.question import (
    Question,
    QuestionCreate,
    QuestionOption,
    QuestionUpdate,
)
from quiz_management.repositories.question_repository import QuestionRepository


def get_utc_now_naive():
    return datetime.now(UTC).replace(tzinfo=None)


def _validate_option_constraints(options: list[dict]) -> None:
    if len(options) != 4:
        raise ServiceException(400, "invalid_payload", "Question must have exactly 4 options")

    if not any(opt["is_correct"] for opt in options):
        raise ServiceException(400, "invalid_payload", "At least one option must be correct")

    indexes = []
    for opt in options:
        if opt["order_index"] is None:
            raise ServiceException(400, "invalid_payload", "Option order index is required")
        if opt["order_index"] < 0:
            raise ServiceException(
                400, "invalid_payload", "Option order index must be non-negative"
            )
        indexes.append(opt["order_index"])

    if len(set(indexes)) != len(indexes):
        raise ServiceException(400, "invalid_payload", "Option order indexes must be unique")


def _prepare_final_options(incoming_options: list, existing_active: dict) -> list[dict]:
    final_options = []
    for opt_in in incoming_options:
        if opt_in.id is not None:
            if opt_in.id not in existing_active:
                raise ServiceException(400, "invalid_payload", "Unknown option id")

            base = existing_active[opt_in.id]
            final_options.append(
                {
                    "id": base.id,
                    "text": opt_in.text if opt_in.text is not None else base.text,
                    "is_correct": opt_in.is_correct
                    if opt_in.is_correct is not None
                    else base.is_correct,
                    "order_index": opt_in.order_index
                    if opt_in.order_index is not None
                    else base.order_index,
                    "is_new": False,
                }
            )
        else:
            if opt_in.text is None:
                raise ServiceException(400, "invalid_payload", "New options must include text")

            final_options.append(
                {
                    "id": None,
                    "text": opt_in.text,
                    "is_correct": opt_in.is_correct or False,
                    "order_index": opt_in.order_index,
                    "is_new": True,
                }
            )
    return final_options


def _apply_options_mutations(
    question: Question, final_options: list[dict], existing_active: dict, now: datetime
) -> None:
    incoming_ids = {opt["id"] for opt in final_options if opt["id"] is not None}
    for opt_id, opt_obj in existing_active.items():
        if opt_id not in incoming_ids:
            opt_obj.deleted_at = now

    for opt_dict in final_options:
        if not opt_dict["is_new"]:
            curr = existing_active[opt_dict["id"]]
            curr.text = opt_dict["text"]
            curr.is_correct = opt_dict["is_correct"]
            curr.order_index = opt_dict["order_index"]
            curr.updated_at = now
        else:
            new_opt = QuestionOption(
                text=opt_dict["text"],
                is_correct=opt_dict["is_correct"],
                order_index=opt_dict["order_index"],
            )
            question.options.append(new_opt)


class QuestionService:
    def __init__(self, db: AsyncSession):
        self.repository = QuestionRepository(db)

    async def _validate_question_order_index_unique(
        self, quiz_id: UUID, order_index: int, exclude_question_id: UUID | None = None
    ) -> None:
        questions = await self.repository.get_by_quiz_id(quiz_id)
        for question in questions:
            if question.order_index == order_index and (
                exclude_question_id is None or question.id != exclude_question_id
            ):
                raise ServiceException(
                    400, "invalid_payload", "Question order_index must be unique within quiz"
                )

    async def get_quiz_questions(self, quiz_id: UUID) -> Sequence[Question]:
        return await self.repository.get_by_quiz_id(quiz_id)

    async def create_question(self, data: QuestionCreate, quiz_id: UUID) -> Question:
        if data.order_index < 0:
            raise ServiceException(
                400, "invalid_payload", "Question order index must be non-negative"
            )

        await self._validate_question_order_index_unique(quiz_id, data.order_index)

        preview_options = [
            {"is_correct": opt.is_correct, "order_index": opt.order_index} for opt in data.options
        ]
        _validate_option_constraints(preview_options)

        sorted_options = sorted(data.options, key=lambda x: x.order_index)
        question = Question(**data.model_dump(exclude={"options"}), quiz_id=quiz_id)
        question.options = []

        for idx, opt_in in enumerate(sorted_options):
            question.options.append(
                QuestionOption(text=opt_in.text, is_correct=opt_in.is_correct, order_index=idx)
            )

        return await self.repository.save(question)

    async def update_question(self, question: Question, data: QuestionUpdate) -> Question:
        now = get_utc_now_naive()

        if data.order_index is not None:
            if data.order_index < 0:
                raise ServiceException(
                    400, "invalid_payload", "Question order index must be non-negative"
                )
            await self._validate_question_order_index_unique(
                question.quiz_id, data.order_index, exclude_question_id=question.id
            )

        if data.options is None:
            data_to_update = data.model_dump(exclude_unset=True, exclude={"options"})
            for key, value in data_to_update.items():
                setattr(question, key, value)
            if data_to_update:
                question.updated_at = now
            return await self.repository.save(question)

        existing_active = {opt.id: opt for opt in question.options if opt.deleted_at is None}
        final_options = _prepare_final_options(data.options, existing_active)

        _validate_option_constraints(final_options)

        final_options.sort(key=lambda x: x["order_index"])
        for new_idx, opt_dict in enumerate(final_options):
            opt_dict["order_index"] = new_idx

        _apply_options_mutations(question, final_options, existing_active, now)

        data_to_update = data.model_dump(exclude_unset=True, exclude={"options"})
        for key, value in data_to_update.items():
            setattr(question, key, value)

        question.updated_at = now
        return await self.repository.save(question)

    async def delete_question(self, question: Question) -> None:
        question.deleted_at = get_utc_now_naive()

        for option in question.options:
            option.deleted_at = get_utc_now_naive()

        await self.repository.save(question)
